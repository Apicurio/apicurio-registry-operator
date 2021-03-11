package patcher

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	ocp_apps "github.com/openshift/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentOCPUF = func(spec *ocp_apps.DeploymentConfig)

type OCPPatcher struct {
	ctx     *context.LoopContext
	clients *client.Clients
}

func NewOCPPatcher(ctx *context.LoopContext, clients *client.Clients) *OCPPatcher {
	return &OCPPatcher{
		ctx:     ctx,
		clients: clients,
	}
}

// ===

func (this *OCPPatcher) reloadDeployment() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_DEPLOYMENT_OCP); exists {
		r, e := this.clients.OCP().
			GetDeployment(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_DEPLOYMENT_OCP)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_DEPLOYMENT_OCP, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	}
}

func (this *OCPPatcher) patchDeployment() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_DEPLOYMENT_OCP,
		func(value interface{}) string {
			return value.(*ocp_apps.DeploymentConfig).String()
		},
		&ocp_apps.DeploymentConfig{},
		"ocp_apps.DeploymentConfig",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			return this.clients.OCP().CreateDeployment(namespace, value.(*ocp_apps.DeploymentConfig))
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.OCP().PatchDeployment(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*ocp_apps.DeploymentConfig).GetName())
		},
	)
}

// =====

func (this *OCPPatcher) reloadRoute() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_ROUTE_OCP); exists {
		r, e := this.clients.OCP().
			GetRoute(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_ROUTE_OCP)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_ROUTE_OCP, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	} else {
		// Load route here, TODO move to separate CF?
		rs, e := this.clients.OCP().
			GetRoutes(this.ctx.GetAppNamespace(), &meta.ListOptions{
				LabelSelector: "app=" + this.ctx.GetAppName().Str(),
			})
		if e == nil {
			existingHost := ""
			if specEntry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SPEC); exists {
				existingHost = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
			}
			for _, r := range rs.Items {
				if r.GetObjectMeta().GetDeletionTimestamp() == nil && r.Spec.Host == existingHost {
					this.ctx.GetResourceCache().Set(resources.RC_KEY_ROUTE_OCP, resources.NewResourceCacheEntry(common.Name(r.Name), &r))
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
