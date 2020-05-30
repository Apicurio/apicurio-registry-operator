package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
)

var _ ControlFunction = &ImageCF{}

// This CF takes care of keeping the "image" section of the CRD applied.
type ImageCF struct {
	ctx              *Context
	deploymentEntry  ResourceCacheEntry
	deploymentExists bool
	existingImage    string
	targetImage      string
}

func NewImageCF(ctx *Context) ControlFunction {
	return &ImageCF{
		ctx:              ctx,
		deploymentEntry:  nil,
		deploymentExists: false,
		existingImage:    RC_EMPTY_NAME,
		targetImage:      RC_EMPTY_NAME,
	}
}

func (this *ImageCF) Describe() string {
	return "ImageCF"
}

func (this *ImageCF) Sense() {
	// Observation #1
	// Get the cached Deployment (if it exists and/or the value)
	deploymentEntry, deploymentExists := this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT)
	this.deploymentEntry = deploymentEntry
	this.deploymentExists = deploymentExists

	// Observation #2
	// Get the existing image name (if present)
	this.existingImage = RC_EMPTY_NAME
	if this.deploymentExists {
		for i, c := range deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetConfiguration().GetAppName() {
				this.existingImage = deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Containers[i].Image
			}
		} // TODO report a problem if not found?
	}

	// Observation #3
	// Get the target image name
	if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
		this.targetImage = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Image.Name
	}
	if this.targetImage == "" {
		// Warning! This is for testing purposes only
		this.targetImage = "apicurio/apicurio-registry-mem:latest-release"
	}

	// Update state
	this.ctx.GetConfiguration().SetConfig(CFG_STA_IMAGE, this.existingImage)
}

func (this *ImageCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Existing image is not the same as the target image (assuming it is never empty)
	return this.deploymentEntry != nil &&
		this.existingImage != this.targetImage
}

func (this *ImageCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetConfiguration().GetAppName() {
				deployment.Spec.Template.Spec.Containers[i].Image = this.targetImage
			}
		} // TODO report a problem if not found?
		return deployment
	})
}
