package cf

import (
	"os"

	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"

	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	apps "k8s.io/api/apps/v1"
)

var _ loop.ControlFunction = &ImageCF{}

const ENV_OPERATOR_REGISTRY_IMAGE_MEM = "REGISTRY_IMAGE_MEM"
const ENV_OPERATOR_REGISTRY_IMAGE_KAFKASQL = "REGISTRY_IMAGE_KAFKASQL"
const ENV_OPERATOR_REGISTRY_IMAGE_SQL = "REGISTRY_IMAGE_SQL"

// This CF takes care of keeping the "image" section of the CRD applied.
type ImageCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcStatus        *status.Status
	deploymentEntry  resources.ResourceCacheEntry
	deploymentExists bool
	existingImage    string
	targetImage      string
	services         *services.LoopServices
	persistence      string
	persistenceError bool
}

func NewImageCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {
	return &ImageCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcStatus:        services.GetStatus(),
		services:         services,
		deploymentEntry:  nil,
		deploymentExists: false,
		existingImage:    "",
		targetImage:      "",
		persistence:      "",
	}
}

func (this *ImageCF) Describe() string {
	return "ImageCF"
}

func (this *ImageCF) Sense() {
	// Observation #1
	// Get the cached Deployment (if it exists and/or the value)
	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)
	this.deploymentEntry = deploymentEntry
	this.deploymentExists = deploymentExists

	// Observation #2
	// Get the existing image name (if present)
	this.existingImage = ""
	if this.deploymentExists {
		for i, c := range deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				this.existingImage = deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Containers[i].Image
			}
		} // TODO report a problem if not found?
	}

	// Observation #3
	// Get the target image name
	this.persistence = ""
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry).Spec
		this.persistence = spec.Configuration.Persistence
		this.targetImage = spec.Deployment.Image
	}

	if this.targetImage == "" {
		envImage := ""
		this.persistenceError = false
		switch this.persistence {
		case "", "mem":
			envImage = os.Getenv(ENV_OPERATOR_REGISTRY_IMAGE_MEM)
		case "kafkasql":
			envImage = os.Getenv(ENV_OPERATOR_REGISTRY_IMAGE_KAFKASQL)
		case "sql":
			envImage = os.Getenv(ENV_OPERATOR_REGISTRY_IMAGE_SQL)
		}
		if envImage != "" {
			this.targetImage = envImage
		} else {
			this.persistenceError = true
			this.ctx.GetLog().WithValues("type", "Warning").
				Info("WARNING: The operand image is not selected. " +
					"Set the 'spec.configuration.persistence' property in your 'apicurioregistry' resource " +
					"to select the appropriate Service Registry image, or set the 'spec.deployment.image' "+
					"property to use a specific image.")
		}
	}

	// Update state
	this.svcStatus.SetConfig(status.CFG_STA_IMAGE, this.existingImage)
}

func (this *ImageCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Existing image is not the same as the target image (assuming it is never empty)
	return (this.deploymentEntry != nil && this.existingImage != this.targetImage) ||
		(this.persistenceError && this.ctx.GetAttempts() == 0)
}

func (this *ImageCF) Respond() {
	if this.persistenceError {
		this.services.GetConditionManager().GetConfigurationErrorCondition().TransitionInvalidPersistence(this.persistence)
		this.services.GetConditionManager().GetReadyCondition().TransitionError()
		return
	}
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				deployment.Spec.Template.Spec.Containers[i].Image = this.targetImage
			}
		} // TODO report a problem if not found?
		return deployment
	})
}

func (this *ImageCF) Cleanup() bool {
	// No cleanup
	return true
}
