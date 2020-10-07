package apicurioregistry

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
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
}

func (this *LabelsCF) Compare() bool {
	this.caLabels = this.GetCommonApplicationLabels()
	this.updateDeployment = this.deploymentIsCached && !labelsEqual(this.deploymentLabels, this.caLabels)
	this.updateDeploymentPod = this.deploymentIsCached && !labelsEqual(this.deploymentPodLabels, this.caLabels)

	return this.podIsCached && (this.updateDeployment || this.updateDeploymentPod)
}

func (this *LabelsCF) Respond() {
	// Response #1
	// Patch Deployment
	if this.updateDeployment {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			labelsUpdate(deployment.Labels, this.caLabels)
			return deployment
		})
	}
	// Response #1
	// Patch Deployment Pod Template
	if this.updateDeploymentPod {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			labelsUpdate(deployment.Spec.Template.Labels, this.caLabels)
			return deployment
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
func labelsEqual(target map[string]string, source map[string]string) bool {
	for sourceKey, sourceValue := range source {
		targetValue, targetExists := target[sourceKey]
		if !targetExists || sourceValue != targetValue {
			return false
		}
	}
	return true
}

func labelsUpdate(target map[string]string, source map[string]string) {
	for sourceKey, sourceValue := range source {
		targetValue, targetExists := target[sourceKey]
		if !targetExists || sourceValue != targetValue {
			target[sourceKey] = sourceValue
		}
	}
}
