package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/status"
	ocp_apps "github.com/openshift/api/apps/v1"
)

var _ loop.ControlFunction = &ReplicasOcpCF{}

type ReplicasOcpCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcStatus        *status.Status
	deploymentEntry  resources.ResourceCacheEntry
	deploymentExists bool
	existingReplicas int32
	targetReplicas   int32
}

// This CF makes sure number of replicas is aligned
// If there is some other way of determining the number of replicas needed outside of CR,
// modify the Sense stage so this CF knows about it
func NewReplicasOcpCF(ctx *context.LoopContext) loop.ControlFunction {
	return &ReplicasOcpCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcStatus:        ctx.GetStatus(),
		deploymentEntry:  nil,
		deploymentExists: false,
		existingReplicas: 0,
		targetReplicas:   0,
	}
}

func (this *ReplicasOcpCF) Describe() string {
	return "ReplicasOcpCF"
}

func (this *ReplicasOcpCF) Sense() {

	// Observation #1
	// Get the cached Deployment (if it exists and/or the value)
	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT_OCP)
	this.deploymentEntry = deploymentEntry
	this.deploymentExists = deploymentExists

	// Observation #2
	// Get the existing replicas (if present)
	this.existingReplicas = 0
	if this.deploymentExists {
		this.existingReplicas = deploymentEntry.GetValue().(*ocp_apps.DeploymentConfig).Spec.Replicas
	}

	// Observation #3
	// Get the target replicas name
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.targetReplicas = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Replicas
	}
	if this.targetReplicas < 1 {
		this.targetReplicas = 1
	}

	// Update state
	this.svcStatus.SetConfigInt32P(status.CFG_STA_REPLICA_COUNT, &this.existingReplicas)
}

func (this *ReplicasOcpCF) Compare() bool {
	// Condition #1
	// Deployment exists
	// Condition #2
	// Existing replicas is not the same as the target replicas (assuming it is never empty)
	return this.deploymentEntry != nil &&
		this.existingReplicas != this.targetReplicas
}

func (this *ReplicasOcpCF) Respond() {
	// Response #1
	// Patch the resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*ocp_apps.DeploymentConfig).DeepCopy()
		deployment.Spec.Replicas = this.targetReplicas
		return deployment
	})
}

func (this *ReplicasOcpCF) Cleanup() bool {
	// No cleanup
	return true
}
