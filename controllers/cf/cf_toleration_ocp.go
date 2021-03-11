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

var _ loop.ControlFunction = &TolerationOcpCF{}

type TolerationOcpCF struct {
	ctx                         *context.LoopContext
	svcResourceCache            resources.ResourceCache
	deploymentConfigEntry       resources.ResourceCacheEntry
	deploymentConfigEntryExists bool
	existingTolerations         []corev1.Toleration
	targetTolerations           []corev1.Toleration
}

func NewTolerationOcpCF(ctx *context.LoopContext) loop.ControlFunction {
	return &TolerationOcpCF{
		ctx:                         ctx,
		svcResourceCache:            ctx.GetResourceCache(),
		deploymentConfigEntry:       nil,
		deploymentConfigEntryExists: false,
		existingTolerations:         nil,
		targetTolerations:           nil,
	}
}

func (this *TolerationOcpCF) Describe() string {
	return "TolerationOcpCF"
}

func (this *TolerationOcpCF) Sense() {
	// Observation #1
	// Get the cached deploymentConfig
	this.deploymentConfigEntry, this.deploymentConfigEntryExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT_OCP)

	if this.deploymentConfigEntryExists {
		// Observation #2
		// Get the existing tolerations
		this.existingTolerations = this.deploymentConfigEntry.GetValue().(*ocp_apps.DeploymentConfig).Spec.Template.Spec.Tolerations

		// Observation #3
		// Get the target tolerations
		if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
			this.targetTolerations = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Tolerations
		}
	}
}

func (this *TolerationOcpCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Target toleration exists
	// Condition #3
	// Existing tolerations are different from target tolerations
	return this.deploymentConfigEntryExists &&
		!reflect.DeepEqual(this.existingTolerations, this.targetTolerations)
}

func (this *TolerationOcpCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentConfigEntry.ApplyPatch(func(value interface{}) interface{} {
		deploymentConfig := value.(*ocp_apps.DeploymentConfig).DeepCopy()
		deploymentConfig.Spec.Template.Spec.Tolerations = this.targetTolerations
		return deploymentConfig
	})
}

func (this *TolerationOcpCF) Cleanup() bool {
	// No cleanup
	return true
}
