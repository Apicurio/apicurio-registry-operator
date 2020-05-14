package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ ControlFunction = &EnvCF{}

type EnvCF struct {
	ctx *Context
}

func NewEnvCF(ctx *Context) ControlFunction {
	return &EnvCF{ctx: ctx}
}

func (this *EnvCF) Describe() string {
	return "Environment Vars Update"
}

func (this *EnvCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	// noop
	return nil
}

func (this *EnvCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	return this.ctx.configuration.EnvChanged(), nil
}

func (this *EnvCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	this.ctx.log.Info("Updating environment variables.")
	this.ctx.patcher.AddDeploymentPatch(func(deployment *apps.Deployment) {
		for i, _ := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Env = this.ctx.configuration.getEnv()
			this.ctx.log.Info("Environment variables updated.")
			return
		}
	})
	return true, nil
}
