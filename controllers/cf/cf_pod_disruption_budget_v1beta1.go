package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	"go.uber.org/zap"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ loop.ControlFunction = &PodDisruptionBudgetV1beta1CF{}

type PodDisruptionBudgetV1beta1CF struct {
	ctx                     context.LoopContext
	log                     *zap.SugaredLogger
	svcResourceCache        resources.ResourceCache
	svcClients              *client.Clients
	svcKubeFactory          *factory.KubeFactory
	svcStatus               *status.Status
	isCached                bool
	podDisruptionBudgets    []policy_v1beta1.PodDisruptionBudget
	podDisruptionBudgetName string
	isPreferred             bool
	disabled                bool
}

func NewPodDisruptionBudgetV1beta1CF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &PodDisruptionBudgetV1beta1CF{
		ctx:                     ctx,
		svcResourceCache:        ctx.GetResourceCache(),
		svcClients:              ctx.GetClients(),
		svcKubeFactory:          services.GetKubeFactory(),
		svcStatus:               services.GetStatus(),
		isCached:                false,
		podDisruptionBudgets:    make([]policy_v1beta1.PodDisruptionBudget, 0),
		podDisruptionBudgetName: resources.RC_NOT_CREATED_NAME_EMPTY,
		isPreferred:             false,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *PodDisruptionBudgetV1beta1CF) Describe() string {
	return "PodDisruptionBudgetV1beta1CF"
}

func (this *PodDisruptionBudgetV1beta1CF) Sense() {
	this.isPreferred = this.ctx.GetSupportedFeatures().PreferredPDBVersion == "v1beta1"

	// Observation #1
	// Get cached PodDisruptionBudget
	pdbEntry, pdbExists := this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1)
	if pdbExists {
		this.podDisruptionBudgetName = pdbEntry.GetName().Str()
	} else {
		this.podDisruptionBudgetName = resources.RC_NOT_CREATED_NAME_EMPTY
	}
	this.isCached = pdbExists

	// Observation #2
	// Get PodDisruptionBudget(s) we *should* track
	this.podDisruptionBudgets = make([]policy_v1beta1.PodDisruptionBudget, 0)
	podDisruptionBudgets, err := this.svcClients.Kube().GetPodDisruptionBudgetsV1beta1(
		this.ctx.GetAppNamespace(),
		meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetAppName().Str(),
		})
	if err == nil {
		for _, podDisruptionBudget := range podDisruptionBudgets.Items {
			// The client will also list v1 resources, and we cannot determine the APIVersion.
			// See https://stackoverflow.com/questions/66757003/list-kubernetes-resources-by-clinet-go-how-can-i-get-the-kind-and-apiversion
			// and https://github.com/kubernetes/client-go/issues/861 .
			// Therefore, as a hack, we will ignore any resources that can also be retrieved using v1 API.
			_, err := this.svcClients.Kube().GetPodDisruptionBudgetV1(
				common.Namespace(podDisruptionBudget.Namespace),
				common.Name(podDisruptionBudget.Name))

			if podDisruptionBudget.GetObjectMeta().GetDeletionTimestamp() == nil &&
				err != nil {
				this.podDisruptionBudgets = append(this.podDisruptionBudgets, podDisruptionBudget)
			}
		}
	}

	this.disabled = false

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := entry.GetValue().(*ar.ApicurioRegistry).Spec
		this.disabled = this.disabled || spec.Deployment.ManagedResources.DisablePodDisruptionBudget
	}

	if this.disabled {
		this.log.Debugw("PodDisruptionBudget is disabled")
	} else {
		this.log.Debugw("PodDisruptionBudget is enabled")
	}

	// Update the status
	if this.isPreferred {
		this.svcStatus.SetConfig(status.CFG_STA_POD_DISRUPTION_BUDGET_NAME, this.podDisruptionBudgetName)
	}
}

func (this *PodDisruptionBudgetV1beta1CF) Compare() bool {
	// Condition #1
	// PodDisruptionBudget is preferred, while cached and at the same time it is disabled (or vice versa)
	// Condition #2
	// If the v1beta1 version is not preferred, we will try to remove it if it exists,
	// so the other CF can create a v1 version instead
	return (this.isPreferred && this.isCached == this.disabled) ||
		(!this.isPreferred && len(this.podDisruptionBudgets) > 0)
}

func (this *PodDisruptionBudgetV1beta1CF) Respond() {
	// Response #1
	// We already know about a PodDisruptionBudget (name), and it is in the list
	if this.podDisruptionBudgetName != resources.RC_NOT_CREATED_NAME_EMPTY {
		contains := false
		for _, val := range this.podDisruptionBudgets {
			if val.Name == this.podDisruptionBudgetName {
				contains = true
				this.svcResourceCache.Set(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1, resources.NewResourceCacheEntry(common.Name(val.Name), &val))
				break
			}
		}
		if !contains {
			this.podDisruptionBudgetName = resources.RC_NOT_CREATED_NAME_EMPTY
		}
	}
	// Response #2
	// Can follow #1, but there must be a single PodDisruptionBudget available
	if this.podDisruptionBudgetName == resources.RC_NOT_CREATED_NAME_EMPTY && len(this.podDisruptionBudgets) == 1 {
		podDisruptionBudget := this.podDisruptionBudgets[0]
		this.podDisruptionBudgetName = podDisruptionBudget.Name
		this.svcResourceCache.Set(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1, resources.NewResourceCacheEntry(common.Name(podDisruptionBudget.Name), &podDisruptionBudget))
	}

	// If this version is not preferred, try to remove it and return
	if !this.isPreferred {
		this.svcResourceCache.Remove(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1)
		for _, v := range this.podDisruptionBudgets {
			if err := this.svcClients.Kube().DeletePodDisruptionBudgetV1beta1(&v); err != nil && !api_errors.IsNotFound(err) {
				this.log.Errorw("could not delete PodDisruptionBudget", "name", v.Name, "error", err)
			}
		}
		return
	}

	// Response #3 (and #4)
	// If there is no service PodDisruptionBudget (or there are more than 1),
	// create a new one IF not disabled
	if !this.disabled && this.podDisruptionBudgetName == resources.RC_NOT_CREATED_NAME_EMPTY && len(this.podDisruptionBudgets) != 1 {
		podDisruptionBudget := this.svcKubeFactory.CreatePodDisruptionBudgetV1beta1()
		// leave the creation itself to patcher+creator so other CFs can update
		this.svcResourceCache.Set(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1, resources.NewResourceCacheEntry(resources.RC_NOT_CREATED_NAME_EMPTY, podDisruptionBudget))
	}

	// Delete an existing PDB if disabled
	// TODO Unify with deletion above
	if this.disabled {
		this.Cleanup()
	}
}

func (this *PodDisruptionBudgetV1beta1CF) Cleanup() bool {
	// PDB should not have any deletion dependencies
	if pdbEntry, pdbExists := this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1); pdbExists {
		if err := this.svcClients.Kube().DeletePodDisruptionBudgetV1beta1(pdbEntry.GetValue().(*policy_v1beta1.PodDisruptionBudget)); err != nil && !api_errors.IsNotFound(err) {
			this.log.Errorw("could not delete PodDisruptionBudget", "error", err)
			return false
		} else {
			this.svcResourceCache.Remove(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1)
			this.ctx.GetLog().Info("PodDisruptionBudget has been deleted")
		}
	}
	return true
}
