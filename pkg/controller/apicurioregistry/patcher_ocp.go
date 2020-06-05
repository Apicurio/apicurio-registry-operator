package apicurioregistry

import (
	ocp_apps "github.com/openshift/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
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

func (this *OCPPatcher) reloadRoute() {
	if entry, exists := this.ctx.GetResourceCache().Get(RC_KEY_ROUTE_OCP); exists {
		r, e := this.ctx.GetClients().OCP().
			GetRoute(this.ctx.GetConfiguration().GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(RC_KEY_ROUTE_OCP)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(RC_KEY_ROUTE_OCP, NewResourceCacheEntry(r.Name, r))
		}
	} else {
		// Load route here, TODO move to separate CF?
		rs, e := this.ctx.GetClients().OCP().
			GetRoutes(this.ctx.GetConfiguration().GetAppNamespace(), &meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetConfiguration().GetAppName(),
		})
		if e == nil {
			existingHost := ""
			if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
				existingHost = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
			}
			for _, r := range rs.Items {
				if r.GetObjectMeta().GetDeletionTimestamp() == nil && r.Spec.Host == existingHost {
					this.ctx.GetResourceCache().Set(RC_KEY_ROUTE_OCP, NewResourceCacheEntry(r.Name, &r))
				}
			}
		}
	}
}

// =====

func (this *OCPPatcher) Reload() {
	this.reloadDeployment()
	this.reloadRoute()
}

func (this *OCPPatcher) Execute() {
	this.patchDeployment()
}
