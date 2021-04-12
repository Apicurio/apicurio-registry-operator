package patcher

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
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

// =====

func (this *OCPPatcher) reloadRoute() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_ROUTE_OCP); exists {
		r, e := this.clients.OCP().
			GetRoute(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Error(e, "Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_ROUTE_OCP)
			this.ctx.SetRequeueNow()
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
	this.reloadRoute()
}

func (this *OCPPatcher) Execute() {
	//empty
}
