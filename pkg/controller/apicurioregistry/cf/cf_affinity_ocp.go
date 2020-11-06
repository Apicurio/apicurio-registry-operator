package cf

import (
	"reflect"

	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	ocp_apps "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ loop.ControlFunction = &AffinityOcpCF{}

type AffinityOcpCF struct {
	ctx                         *context.LoopContext
	svcResourceCache            resources.ResourceCache
	deploymentConfigEntry       resources.ResourceCacheEntry
	deploymentConfigEntryExists bool
	existingAffinity            *corev1.Affinity
	targetAffinity              *corev1.Affinity
}

func NewAffinityOcpCF(ctx *context.LoopContext) loop.ControlFunction {
	return &AffinityOcpCF{
		ctx:                         ctx,
		svcResourceCache:            ctx.GetResourceCache(),
		deploymentConfigEntry:       nil,
		deploymentConfigEntryExists: false,
		existingAffinity:            nil,
		targetAffinity:              nil,
	}
}

func (this *AffinityOcpCF) Describe() string {
	return "AffinityOcpCF"
}

func (this *AffinityOcpCF) Sense() {
	// Observation #1
	// Get the cached deploymentConfig
	this.deploymentConfigEntry, this.deploymentConfigEntryExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT_OCP)

	if this.deploymentConfigEntryExists {
		// Observation #2
		// Get the existing affinity
		this.existingAffinity = this.deploymentConfigEntry.GetValue().(*ocp_apps.DeploymentConfig).Spec.Template.Spec.Affinity

		// Observation #3
		// Get the target affinity
		if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
			this.targetAffinity = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Affinity
		}
	}
}

func (this *AffinityOcpCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target affinity exists
	// Condition #3
	// Existing affinity is different from target affinity
	return this.deploymentConfigEntryExists &&
		!reflect.DeepEqual(this.existingAffinity, this.targetAffinity)
}

func (this *AffinityOcpCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentConfigEntry.ApplyPatch(func(value interface{}) interface{} {
		deploymentConfig := value.(*ocp_apps.DeploymentConfig).DeepCopy()
		deploymentConfig.Spec.Template.Spec.Affinity = this.targetAffinity
		return deploymentConfig
	})
}

func (this *AffinityOcpCF) Cleanup() bool {
	// No cleanup
	return true
}
