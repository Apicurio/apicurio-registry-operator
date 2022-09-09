package cf

import (
	"reflect"

	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"

	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ loop.ControlFunction = &ImagePullSecretsCF{}

type ImagePullSecretsCF struct {
	ctx                      context.LoopContext
	svcResourceCache         resources.ResourceCache
	deploymentEntry          resources.ResourceCacheEntry
	deploymentEntryExists    bool
	existingImagePullSecrets []core.LocalObjectReference
	targetImagePullSecrets   []core.LocalObjectReference
}

func NewImagePullSecretsCF(ctx context.LoopContext) loop.ControlFunction {
	return &ImagePullSecretsCF{
		ctx:                      ctx,
		svcResourceCache:         ctx.GetResourceCache(),
		deploymentEntry:          nil,
		deploymentEntryExists:    false,
		existingImagePullSecrets: nil,
		targetImagePullSecrets:   nil,
	}
}

func (this *ImagePullSecretsCF) Describe() string {
	return "ImagePullSecretsCF"
}

func (this *ImagePullSecretsCF) Sense() {
	// Observation #1
	// Get the cached deployment
	this.deploymentEntry, this.deploymentEntryExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)

	if this.deploymentEntryExists {
		// Observation #2
		// Get the existing pod ImagePullSecrets
		this.existingImagePullSecrets = this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.ImagePullSecrets

		// Observation #3
		// Get the target pod ImagePullSecrets
		if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
			this.targetImagePullSecrets = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.ImagePullSecrets
		}
	}
}

func (this *ImagePullSecretsCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target pod ImagePullSecrets exists
	// Condition #3
	// Existing pod ImagePullSecrets is different to target pod ImagePullSecrets
	return this.deploymentEntryExists &&
		!reflect.DeepEqual(this.existingImagePullSecrets, this.targetImagePullSecrets)
}

func (this *ImagePullSecretsCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		deployment.Spec.Template.Spec.ImagePullSecrets = this.targetImagePullSecrets
		return deployment
	})
}

func (this *ImagePullSecretsCF) Cleanup() bool {
	// No cleanup
	return true
}
