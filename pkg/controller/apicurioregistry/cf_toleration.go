package apicurioregistry

import (
	"reflect"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ ControlFunction = &TolerationCF{}

type TolerationCF struct {
	ctx                   *Context
	deploymentEntry       ResourceCacheEntry
	deploymentEntryExists bool
	existingTolerations   []corev1.Toleration
	targetTolerations     []corev1.Toleration
}

func NewTolerationCF(ctx *Context) ControlFunction {
	return &TolerationCF{
		ctx:                   ctx,
		deploymentEntry:       nil,
		deploymentEntryExists: false,
		existingTolerations:   nil,
		targetTolerations:     nil,
	}
}

func (this *TolerationCF) Describe() string {
	return "TolerationCF"
}

func (this *TolerationCF) Sense() {
	// Observation #1
	// Get the cached deployment
	this.deploymentEntry, this.deploymentEntryExists = this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT)

	if this.deploymentEntryExists {
		// Observation #2
		// Get the existing tolerations
		this.existingTolerations = this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Tolerations

		// Observation #3
		// Get the target tolerations
		if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
			this.targetTolerations = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Tolerations
		}
	}
}

func (this *TolerationCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target toleration exists
	// Condition #3
	// Existing tolerations are different from target tolerations
	return this.deploymentEntryExists &&
		!reflect.DeepEqual(this.existingTolerations, this.targetTolerations)
}

func (this *TolerationCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		deployment.Spec.Template.Spec.Tolerations = this.targetTolerations
		return deployment
	})
}
