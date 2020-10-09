package apicurioregistry

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	policy "k8s.io/api/policy/v1beta1"
)

var _ ControlFunction = &LabelsCF{}

type LabelsCF struct {
	ctx *Context

	podEntry    ResourceCacheEntry
	podIsCached bool
	podLabels   map[string]string

	caLabels map[string]string

	deploymentEntry     ResourceCacheEntry
	deploymentIsCached  bool
	deploymentLabels    map[string]string
	deploymentPodLabels map[string]string
	updateDeployment    bool
	updateDeploymentPod bool

	serviceEntry    ResourceCacheEntry
	serviceIsCached bool
	serviceLabels   map[string]string
	updateService   bool

	ingressEntry    ResourceCacheEntry
	ingressIsCached bool
	ingressLabels   map[string]string
	updateIngress   bool

	pdbEntry    ResourceCacheEntry
	pdbIsCached bool
	pdbLabels   map[string]string
	updatePdb   bool
}

// Update labels on some managed resources
func NewLabelsCF(ctx *Context) ControlFunction {
	return &LabelsCF{
		ctx:       ctx,
		podLabels: nil,
	}
}

func (this *LabelsCF) Describe() string {
	return "LabelsCF"
}

func (this *LabelsCF) Sense() {
	// Observation #1
	// Operator Pod
	this.podEntry, this.podIsCached = this.ctx.GetResourceCache().Get(RC_KEY_OPERATOR_POD)
	if this.podIsCached {
		this.podLabels = this.podEntry.GetValue().(*core.Pod).Labels
	}
	// Observation #2
	// Deployment & Deployment Pod Template
	this.deploymentEntry, this.deploymentIsCached = this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT)
	if this.deploymentIsCached {
		this.deploymentLabels = this.deploymentEntry.GetValue().(*apps.Deployment).Labels
		this.deploymentPodLabels = this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Labels
	}
	// Observation #3
	// Service
	this.serviceEntry, this.serviceIsCached = this.ctx.GetResourceCache().Get(RC_KEY_SERVICE)
	if this.serviceIsCached {
		this.serviceLabels = this.serviceEntry.GetValue().(*core.Service).Labels
	}
	// Observation #4
	// Ingress
	this.ingressEntry, this.ingressIsCached = this.ctx.GetResourceCache().Get(RC_KEY_INGRESS)
	if this.ingressIsCached {
		this.ingressLabels = this.ingressEntry.GetValue().(*extensions.Ingress).Labels
	}
	// Observation #5
	// PodDisruptionBudget
	this.pdbEntry, this.pdbIsCached = this.ctx.GetResourceCache().Get(RC_KEY_POD_DISRUPTION_BUDGET)
	if this.pdbIsCached {
		this.pdbLabels = this.pdbEntry.GetValue().(*policy.PodDisruptionBudget).Labels
	}
}

func (this *LabelsCF) Compare() bool {
	this.caLabels = this.GetCommonApplicationLabels()
	this.updateDeployment = this.deploymentIsCached && !LabelsEqual(this.deploymentLabels, this.caLabels)
	this.updateDeploymentPod = this.deploymentIsCached && !LabelsEqual(this.deploymentPodLabels, this.caLabels)
	this.updateService = this.serviceIsCached && !LabelsEqual(this.serviceLabels, this.caLabels)
	this.updateIngress = this.ingressIsCached && !LabelsEqual(this.ingressLabels, this.caLabels)
	this.updatePdb = this.pdbIsCached && !LabelsEqual(this.pdbLabels, this.caLabels)

	return this.podIsCached && (
		this.updateDeployment ||
			this.updateDeploymentPod ||
			this.updateService ||
			this.updateIngress ||
			this.updatePdb)
}

func (this *LabelsCF) Respond() {
	// Response #1
	// Patch Deployment
	if this.updateDeployment {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			LabelsUpdate(deployment.Labels, this.caLabels)
			return deployment
		})
	}
	// Response #2
	// Patch Deployment Pod Template
	if this.updateDeploymentPod {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			LabelsUpdate(deployment.Spec.Template.Labels, this.caLabels)
			return deployment
		})
	}
	// Response #3
	// Service
	if this.updateService {
		this.serviceEntry.ApplyPatch(func(value interface{}) interface{} {
			service := value.(*core.Service).DeepCopy()
			LabelsUpdate(service.Labels, this.caLabels)
			return service
		})
	}
	// Response #4
	// Ingress
	if this.updateIngress {
		this.ingressEntry.ApplyPatch(func(value interface{}) interface{} {
			ingress := value.(*extensions.Ingress).DeepCopy()
			LabelsUpdate(ingress.Labels, this.caLabels)
			return ingress
		})
	}
	// Response #5
	// PodDisruptionBudget
	if this.updatePdb {
		this.pdbEntry.ApplyPatch(func(value interface{}) interface{} {
			pdb := value.(*policy.PodDisruptionBudget).DeepCopy()
			LabelsUpdate(pdb.Labels, this.caLabels)
			return pdb
		})
	}
}

func (this *LabelsCF) Cleanup() bool {
	// No cleanup
	return true
}

// ---

func (this *LabelsCF) GetCommonApplicationLabels() map[string]string {
	return this.ctx.GetKubeFactory().GetLabels()
}

// Return *true* if, for given source labels,
// the target label values exist and have the same value
func LabelsEqual(target map[string]string, source map[string]string) bool {
	for sourceKey, sourceValue := range source {
		targetValue, targetExists := target[sourceKey]
		if !targetExists || sourceValue != targetValue {
			return false
		}
	}
	return true
}

func LabelsUpdate(target map[string]string, source map[string]string) {
	for sourceKey, sourceValue := range source {
		targetValue, targetExists := target[sourceKey]
		if !targetExists || sourceValue != targetValue {
			target[sourceKey] = sourceValue
		}
	}
}
