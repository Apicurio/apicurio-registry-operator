package apicurioregistry

import (
	ar "github.com/apicurio/apicurio-operators/apicurio-registry/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ ControlFunction = &ConfReplicasCF{}

type ConfReplicasCF struct {
	ctx *Context
}

// This CF makes sure number of replicas is aligned
// If there is some other way of determining the number of replicas needed outside of CR,
// modify the Sense stage so this CF knows about it
func NewConfReplicasCF(ctx *Context) ControlFunction {
	return &ConfReplicasCF{ctx: ctx}
}

func (this *ConfReplicasCF) Describe() string {
	return "Align configured replica count"
}

func (this *ConfReplicasCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	log := this.ctx.log.WithValues("CF", this.Describe())

	if deployment, err := this.ctx.kubecl.GetDeployment(); err == nil {
		this.ctx.configuration.SetConfigInt32P(CFG_STA_REPLICA_COUNT, deployment.Spec.Replicas)
	} else {
		log.Info("Warn: Could not get deployment.")
		return err
	}
	return nil
}

func (this *ConfReplicasCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {

	return /* this.ctx.configuration.GetConfig(CFG_STA_REPLICA_COUNT) != "" && */ this.
		ctx.configuration.GetConfig(CFG_STA_REPLICA_COUNT) != this.ctx.configuration.GetConfig(CFG_DEP_REPLICAS), nil
}

func (this *ConfReplicasCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {

	this.ctx.patcher.AddDeploymentPatch(func(deployment *apps.Deployment) {
		deployment.Spec.Replicas = this.ctx.configuration.GetConfigInt32P(CFG_DEP_REPLICAS)
	})
	return true, nil
}
