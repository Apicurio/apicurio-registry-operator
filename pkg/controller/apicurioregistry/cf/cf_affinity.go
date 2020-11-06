package cf

import (
	"reflect"

	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ loop.ControlFunction = &AffinityCF{}

type AffinityCF struct {
	ctx                   *context.LoopContext
	svcResourceCache      resources.ResourceCache
	deploymentEntry       resources.ResourceCacheEntry
	deploymentEntryExists bool
	existingAffinity      *corev1.Affinity
	targetAffinity        *corev1.Affinity
}

func NewAffinityCF(ctx *context.LoopContext) loop.ControlFunction {
	return &AffinityCF{
		ctx:                   ctx,
		svcResourceCache:      ctx.GetResourceCache(),
		deploymentEntry:       nil,
		deploymentEntryExists: false,
		existingAffinity:      nil,
		targetAffinity:        nil,
	}
}

func (this *AffinityCF) Describe() string {
	return "AffinityCF"
}

func (this *AffinityCF) Sense() {
	// Observation #1
	// Get the cached deployment
	this.deploymentEntry, this.deploymentEntryExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)

	if this.deploymentEntryExists {
		// Observation #2
		// Get the existing affinity
		this.existingAffinity = this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Spec.Affinity

		// Observation #3
		// Get the target affinity
		if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
			this.targetAffinity = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Affinity
		}
	}
}

func (this *AffinityCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target affinity exists
	// Condition #3
	// Existing affinity is different from target affinity
	return this.deploymentEntryExists &&
		!reflect.DeepEqual(this.existingAffinity, this.targetAffinity)
}

func (this *AffinityCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		deployment.Spec.Template.Spec.Affinity = this.targetAffinity
		return deployment
	})
}

func (this *AffinityCF) Cleanup() bool {
	// No cleanup
	return true
}
