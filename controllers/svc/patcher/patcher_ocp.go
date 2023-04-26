package patcher

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	ocp_apps "github.com/openshift/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentOCPUF = func(spec *ocp_apps.DeploymentConfig)

type OCPPatcher struct {
	ctx context.LoopContext
}

func NewOCPPatcher(ctx context.LoopContext) *OCPPatcher {
	return &OCPPatcher{
		ctx,
	}
}

// =====

func (this *OCPPatcher) reloadRoute() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_ROUTE_OCP); exists {
		r, e := this.ctx.GetClients().OCP().
			GetRoute(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
				"name", entry.GetName(), "error", e)
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_ROUTE_OCP)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_ROUTE_OCP, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	} else {
		// Load route here, TODO move to separate CF?
		rs, e := this.ctx.GetClients().OCP().
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
