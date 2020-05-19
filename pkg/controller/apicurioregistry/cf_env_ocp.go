package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	ocp_apps "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ ControlFunction = &EnvOCPCF{}

type EnvOCPCF struct {
	ctx *Context
}

func NewEnvOCPCF(ctx *Context) ControlFunction {
	return &EnvOCPCF{ctx: ctx}
}

func (this *EnvOCPCF) Describe() string {
	return "Environment Vars Update (OCP)"
}

func (this *EnvOCPCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	// noop
	return nil
}

func (this *EnvOCPCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	return this.ctx.GetConfiguration().EnvChanged(), nil
}

func (this *EnvOCPCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	this.ctx.GetLog().Info("Updating environment variables.")
	this.ctx.GetPatchers().OCP().AddDeploymentPatch(func(deployment *ocp_apps.DeploymentConfig) {
		for i, _ := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Env = this.ctx.GetConfiguration().GetEnv()
			this.ctx.GetLog().Info("Environment variables updated.")
			return
		}
	})
	return true, nil
}
