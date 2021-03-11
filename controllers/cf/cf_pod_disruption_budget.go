package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	policy "k8s.io/api/policy/v1beta1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ loop.ControlFunction = &PodDisruptionBudgetCF{}

type PodDisruptionBudgetCF struct {
	ctx                     *context.LoopContext
	svcResourceCache        resources.ResourceCache
	svcClients              *client.Clients
	svcKubeFactory          *factory.KubeFactory
	isCached                bool
	podDisruptionBudgets    []policy.PodDisruptionBudget
	podDisruptionBudgetName string
}

func NewPodDisruptionBudgetCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {

	return &PodDisruptionBudgetCF{
		ctx:                     ctx,
		svcResourceCache:        ctx.GetResourceCache(),
		svcClients:              services.GetClients(),
		svcKubeFactory:          services.GetKubeFactory(),
		isCached:                false,
		podDisruptionBudgets:    make([]policy.PodDisruptionBudget, 0),
		podDisruptionBudgetName: resources.RC_EMPTY_NAME,
	}
}

func (this *PodDisruptionBudgetCF) Describe() string {
	return "PodDisruptionBudgetCF"
}

func (this *PodDisruptionBudgetCF) Sense() {

	// Observation #1
	// Get cached PodDisruptionBudget
	pdbEntry, pdbExists := this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET)
	if pdbExists {
		this.podDisruptionBudgetName = pdbEntry.GetName().Str()
	} else {
		this.podDisruptionBudgetName = resources.RC_EMPTY_NAME
	}
	this.isCached = pdbExists

	// Observation #2
	// Get PodDisruptionBudget(s) we *should* track
	this.podDisruptionBudgets = make([]policy.PodDisruptionBudget, 0)
	podDisruptionBudgets, err := this.svcClients.Kube().GetPodDisruptionBudgets(
		this.ctx.GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetAppName().Str(),
		})
	if err == nil {
		for _, podDisruptionBudget := range podDisruptionBudgets.Items {
			if podDisruptionBudget.GetObjectMeta().GetDeletionTimestamp() == nil {
				this.podDisruptionBudgets = append(this.podDisruptionBudgets, podDisruptionBudget)
			}
		}
	}
}

func (this *PodDisruptionBudgetCF) Compare() bool {
	// Condition #1
	// If we already have a PodDisruptionBudget cached, skip
	return !this.isCached
}

func (this *PodDisruptionBudgetCF) Respond() {
	// Response #1
	// We already know about a PodDisruptionBudget (name), and it is in the list
	if this.podDisruptionBudgetName != resources.RC_EMPTY_NAME {
		contains := false
		for _, val := range this.podDisruptionBudgets {
			if val.Name == this.podDisruptionBudgetName {
				contains = true
				this.svcResourceCache.Set(resources.RC_KEY_POD_DISRUPTION_BUDGET, resources.NewResourceCacheEntry(common.Name(val.Name), &val))
				break
			}
		}
		if !contains {
			this.podDisruptionBudgetName = resources.RC_EMPTY_NAME
		}
	}
	// Response #2
	// Can follow #1, but there must be a single PodDisruptionBudget available
	if this.podDisruptionBudgetName == resources.RC_EMPTY_NAME && len(this.podDisruptionBudgets) == 1 {
		podDisruptionBudget := this.podDisruptionBudgets[0]
		this.podDisruptionBudgetName = podDisruptionBudget.Name
		this.svcResourceCache.Set(resources.RC_KEY_POD_DISRUPTION_BUDGET, resources.NewResourceCacheEntry(common.Name(podDisruptionBudget.Name), &podDisruptionBudget))
	}
	// Response #3 (and #4)
	// If there is no service PodDisruptionBudget (or there are more than 1), just create a new one
	if this.podDisruptionBudgetName == resources.RC_EMPTY_NAME && len(this.podDisruptionBudgets) != 1 {
		podDisruptionBudget := this.svcKubeFactory.CreatePodDisruptionBudget()
		// leave the creation itself to patcher+creator so other CFs can update
		this.svcResourceCache.Set(resources.RC_KEY_POD_DISRUPTION_BUDGET, resources.NewResourceCacheEntry(resources.RC_EMPTY_NAME, podDisruptionBudget))
	}
}

func (this *PodDisruptionBudgetCF) Cleanup() bool {
	// PDB should not have any deletion dependencies
	if pdbEntry, pdbExists := this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET); pdbExists {
		if err := this.svcClients.Kube().DeletePodDisruptionBudget(pdbEntry.GetValue().(*policy.PodDisruptionBudget), &meta.DeleteOptions{}); err != nil && !api_errors.IsNotFound(err) {
			this.ctx.GetLog().Error(err, "Could not delete PodDisruptionBudget during cleanup")
			return false
		} else {
			this.svcResourceCache.Remove(resources.RC_KEY_POD_DISRUPTION_BUDGET)
			this.ctx.GetLog().Info("PodDisruptionBudget has been deleted.")
		}
	}
	return true
}
