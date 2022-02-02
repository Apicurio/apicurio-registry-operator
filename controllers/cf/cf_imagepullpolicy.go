package cf

import (
	"os"

	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ loop.ControlFunction = &ImagePullPolicyCF{}

const ENV_OPERATOR_REGISTRY_IMAGE_PULL_POLICY = "REGISTRY_IMAGE_PULL_POLICY"

type ImagePullPolicyCF struct {
	ctx                     *context.LoopContext
	svcResourceCache        resources.ResourceCache
	deploymentEntry         resources.ResourceCacheEntry
	deploymentEntryExists   bool
	existingImagePullPolicy core.PullPolicy
	targetImagePullPolicy   core.PullPolicy
}

func NewImagePullPolicyCF(ctx *context.LoopContext) loop.ControlFunction {
	return &ImagePullPolicyCF{
		ctx:                     ctx,
		svcResourceCache:        ctx.GetResourceCache(),
		deploymentEntry:         nil,
		deploymentEntryExists:   false,
		existingImagePullPolicy: "",
		targetImagePullPolicy:   "",
	}
}

func (this *ImagePullPolicyCF) Describe() string {
	return "ImagePullPolicyCF"
}

func (this *ImagePullPolicyCF) Sense() {
	// Observation #1
	// Get the cached deployment
	this.deploymentEntry, this.deploymentEntryExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)

	if this.deploymentEntryExists {
		// Observation #2
		// Get the existing pod ImagePullPolicy
		for i, c := range this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				this.existingImagePullPolicy = this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Containers[i].ImagePullPolicy
			}
		}

		// Observation #3
		// Get the target pod imagePullPolicy from the REGISTRY_IMAGE_PULL_POLICY env variable
		envImagePullPolicy := os.Getenv(ENV_OPERATOR_REGISTRY_IMAGE_PULL_POLICY)
		switch envImagePullPolicy {
		case string(core.PullAlways):
			this.targetImagePullPolicy = core.PullAlways
		case string(core.PullNever):
			this.targetImagePullPolicy = core.PullNever
		case string(core.PullIfNotPresent):
			this.targetImagePullPolicy = core.PullIfNotPresent
		}

		if envImagePullPolicy != "" && this.targetImagePullPolicy == "" {
			this.ctx.GetLog().WithValues("type", "Warning").
				Info("WARNING: " + envImagePullPolicy + " is not a valid value for " + ENV_OPERATOR_REGISTRY_IMAGE_PULL_POLICY + ". " +
					ENV_OPERATOR_REGISTRY_IMAGE_PULL_POLICY + " can have one of the following values: Always, IfNotPresent, Never.")
		}
	}
}

func (this *ImagePullPolicyCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target pod imagePullPolicy exists
	// Condition #3
	// Existing pod imagePullPolicy is different to target pod imagePullPolicy
	return this.deploymentEntryExists && this.targetImagePullPolicy != "" &&
		this.existingImagePullPolicy != this.targetImagePullPolicy
}

func (this *ImagePullPolicyCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				deployment.Spec.Template.Spec.Containers[i].ImagePullPolicy = this.targetImagePullPolicy
			}
		}
		return deployment
	})
}

func (this *ImagePullPolicyCF) Cleanup() bool {
	// No cleanup
	return true
}
