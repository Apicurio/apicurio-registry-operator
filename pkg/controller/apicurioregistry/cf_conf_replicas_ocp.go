package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	ocp_apps "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ ControlFunction = &ConfReplicasOCPCF{}

type ConfReplicasOCPCF struct {
	ctx *Context
}

// This CF makes sure number of replicas is aligned
// If there is some other way of determining the number of replicas needed outside of CR,
// modify the Sense stage so this CF knows about it
func NewConfReplicasOCPCF(ctx *Context) ControlFunction {
	return &ConfReplicasOCPCF{ctx: ctx}
}

func (this *ConfReplicasOCPCF) Describe() string {
	return "Align configured replica count (OCP)"
}

func (this *ConfReplicasOCPCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	log := this.ctx.GetLog().WithValues("CF", this.Describe())

	if deployment, err := this.ctx.GetClients().OCP().GetCurrentDeployment(); err == nil {
		this.ctx.GetConfiguration().SetConfigInt32P(CFG_STA_REPLICA_COUNT, &deployment.Spec.Replicas)
	} else {
		log.Info("Warn: Could not get deployment.")
		return err
	}
	return nil
}

func (this *ConfReplicasOCPCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {

	return /* this.ctx.GetConfiguration().GetConfig(CFG_STA_REPLICA_COUNT) != "" && */ this.
		ctx.GetConfiguration().GetConfig(CFG_STA_REPLICA_COUNT) != this.ctx.GetConfiguration().GetConfig(CFG_DEP_REPLICAS), nil
}

func (this *ConfReplicasOCPCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {

	this.ctx.GetPatchers().OCP().AddDeploymentPatch(func(deployment *ocp_apps.DeploymentConfig) {
		deployment.Spec.Replicas = *this.ctx.GetConfiguration().GetConfigInt32P(CFG_DEP_REPLICAS)
	})
	return true, nil
}
