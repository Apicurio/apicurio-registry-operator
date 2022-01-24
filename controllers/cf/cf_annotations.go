package cf

import (
	"reflect"

	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"

	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	apps "k8s.io/api/apps/v1"
)

var _ loop.ControlFunction = &AnnotationsCF{}

type AnnotationsCF struct {
	ctx                   *context.LoopContext
	svcResourceCache      resources.ResourceCache
	deploymentEntry       resources.ResourceCacheEntry
	deploymentEntryExists bool
	existingAnnotations   map[string]string
	targetAnnotations     map[string]string
}

func NewAnnotationsCF(ctx *context.LoopContext) loop.ControlFunction {
	return &AnnotationsCF{
		ctx:                   ctx,
		svcResourceCache:      ctx.GetResourceCache(),
		deploymentEntry:       nil,
		deploymentEntryExists: false,
		existingAnnotations:   nil,
		targetAnnotations:     nil,
	}
}

func (this *AnnotationsCF) Describe() string {
	return "AnnotationsCF"
}

func (this *AnnotationsCF) Sense() {
	// Observation #1
	// Get the cached deployment
	this.deploymentEntry, this.deploymentEntryExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)

	if this.deploymentEntryExists {
		// Observation #2
		// Get the existing pod annotations
		this.existingAnnotations = this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Annotations

		// Observation #3
		// Get the target pod annotations
		if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
			this.targetAnnotations = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Metadata.Annotations
		}
	}
}

func (this *AnnotationsCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target pod annotations exists
	// Condition #3
	// Existing pod annotations are different to target pod annotations
	return this.deploymentEntryExists &&
		!reflect.DeepEqual(this.existingAnnotations, this.targetAnnotations)
}

func (this *AnnotationsCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		deployment.Spec.Template.Annotations = this.targetAnnotations
		return deployment
	})
}

func (this *AnnotationsCF) Cleanup() bool {
	// No cleanup
	return true
}
