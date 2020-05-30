package apicurioregistry

import (
	ocp_apps "github.com/openshift/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentOCPUF = func(spec *ocp_apps.DeploymentConfig)

type OCPPatcher struct {
	ctx           *Context
	deploymentUFs []DeploymentOCPUF
}

func NewOCPPatcher(ctx *Context) *OCPPatcher {
	return &OCPPatcher{
		ctx: ctx,
	}
}

// ===

func (this *OCPPatcher) reloadDeployment() {
	if entry, exists := this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT_OCP); exists {
		r, e := this.ctx.GetClients().OCP().
			GetDeployment(this.ctx.GetConfiguration().GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(RC_KEY_DEPLOYMENT_OCP)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(RC_KEY_DEPLOYMENT_OCP, NewResourceCacheEntry(r.Name, r))
		}
	}
}

func (this *OCPPatcher) patchDeployment() {
	patchGeneric(
		this.ctx,
		RC_KEY_DEPLOYMENT_OCP,
		func(namespace string, name string) (interface{}, error) {
			return this.ctx.GetClients().OCP().GetDeployment(namespace, name, &meta.GetOptions{})
		},
		func(value interface{}) string {
			return value.(*ocp_apps.DeploymentConfig).String()
		},
		&ocp_apps.DeploymentConfigSpec{},
		"ocp_apps.DeploymentConfig",
		func(namespace string, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().OCP().CreateDeployment(namespace, value.(*ocp_apps.DeploymentConfig))
		},
		func(namespace string, name string, data []byte) (interface{}, error) {
			return this.ctx.GetClients().OCP().PatchDeployment(namespace, name, data)
		},
		func(value interface{}) string {
			return value.(*ocp_apps.DeploymentConfig).GetName()
		},
		func(value interface{}) interface{} {
			return value.(*ocp_apps.DeploymentConfig).Spec
		},
	)
}

// =====

func (this *OCPPatcher) Reload() {
	this.reloadDeployment()
}

func (this *OCPPatcher) Execute() {
	this.patchDeployment()
}
