package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &PodDisruptionBudgetCF{}

type PodDisruptionBudgetCF struct {
	ctx                     *Context
	isCached                bool
	podDisruptionBudgets    []policy.PodDisruptionBudget
	podDisruptionBudgetName string
}

func NewPodDisruptionBudgetCF(ctx *Context) ControlFunction {

	err := ctx.GetController().Watch(&source.Kind{Type: &policy.PodDisruptionBudget{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating PodDisruptionBudget watch.")
	}

	return &PodDisruptionBudgetCF{
		ctx:                     ctx,
		isCached:                false,
		podDisruptionBudgets:    make([]policy.PodDisruptionBudget, 0),
		podDisruptionBudgetName: RC_EMPTY_NAME,
	}
}

func (this *PodDisruptionBudgetCF) Describe() string {
	return "PodDisruptionBudgetCF"
}

func (this *PodDisruptionBudgetCF) Sense() {

	// Observation #1
	// Get cached PodDisruptionBudget
	pdbEntry, pdbExists := this.ctx.GetResourceCache().Get(RC_KEY_POD_DISRUPTION_BUDGET)
	if pdbExists {
		this.podDisruptionBudgetName = pdbEntry.GetName()
	} else {
		this.podDisruptionBudgetName = RC_EMPTY_NAME
	}
	this.isCached = pdbExists

	// Observation #2
	// Get PodDisruptionBudget(s) we *should* track
	this.podDisruptionBudgets = make([]policy.PodDisruptionBudget, 0)
	podDisruptionBudgets, err := this.ctx.GetClients().Kube().GetPodDisruptionBudgets(
		this.ctx.GetConfiguration().GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetConfiguration().GetAppName(),
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
	if this.podDisruptionBudgetName != RC_EMPTY_NAME {
		contains := false
		for _, val := range this.podDisruptionBudgets {
			if val.Name == this.podDisruptionBudgetName {
				contains = true
				this.ctx.GetResourceCache().Set(RC_KEY_POD_DISRUPTION_BUDGET, NewResourceCacheEntry(val.Name, &val))
				break
			}
		}
		if !contains {
			this.podDisruptionBudgetName = RC_EMPTY_NAME
		}
	}
	// Response #2
	// Can follow #1, but there must be a single PodDisruptionBudget available
	if this.podDisruptionBudgetName == RC_EMPTY_NAME && len(this.podDisruptionBudgets) == 1 {
		podDisruptionBudget := this.podDisruptionBudgets[0]
		this.podDisruptionBudgetName = podDisruptionBudget.Name
		this.ctx.GetResourceCache().Set(RC_KEY_POD_DISRUPTION_BUDGET, NewResourceCacheEntry(podDisruptionBudget.Name, &podDisruptionBudget))
	}
	// Response #3 (and #4)
	// If there is no service PodDisruptionBudget (or there are more than 1), just create a new one
	if this.podDisruptionBudgetName == RC_EMPTY_NAME && len(this.podDisruptionBudgets) != 1 {
		podDisruptionBudget := this.ctx.GetKubeFactory().CreatePodDisruptionBudget()
		// leave the creation itself to patcher+creator so other CFs can update
		this.ctx.GetResourceCache().Set(RC_KEY_POD_DISRUPTION_BUDGET, NewResourceCacheEntry(RC_EMPTY_NAME, podDisruptionBudget))
	}
}
