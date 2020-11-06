package patcher

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubePatcher struct {
	ctx     *context.LoopContext
	clients *client.Clients
}

func NewKubePatcher(ctx *context.LoopContext, clients *client.Clients) *KubePatcher {
	return &KubePatcher{
		ctx:     ctx,
		clients: clients,
	}
}

// ===

func (this *KubePatcher) reloadApicurioRegistry() {
	// No need to check if the entry exists
	r, e := this.clients.CRD().GetApicurioRegistry(this.ctx.GetAppNamespace(), this.ctx.GetAppName(), &meta.GetOptions{})
	if e != nil {
		this.ctx.GetLog().WithValues("name", this.ctx.GetAppName()).Info("Resource not found. (May have been deleted).")
		this.ctx.GetResourceCache().Remove(resources.RC_KEY_SPEC)
		this.ctx.SetRequeue()
	} else {
		this.ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(common.Name(r.Name), r))
	}
}

func (this *KubePatcher) patchApicurioRegistry() { // TODO move to separate file/class?
	patchGeneric(
		this.ctx,
		resources.RC_KEY_SPEC,
		func(value interface{}) string {
			return value.(*ar.ApicurioRegistry).ObjectMeta.String()
		},
		&ar.ApicurioRegistry{},
		"ar.ApicurioRegistry",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			// This should be not used (at the moment)
			panic("Unsupported operation.")
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.CRD().PatchApicurioRegistry(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*ar.ApicurioRegistry).GetName())
		},
	)
}

func (this *KubePatcher) reloadDeployment() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_DEPLOYMENT); exists {
		r, e := this.clients.Kube().GetDeployment(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_DEPLOYMENT)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_DEPLOYMENT, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	}
}

func (this *KubePatcher) patchDeployment() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_DEPLOYMENT,
		func(value interface{}) string {
			return value.(*apps.Deployment).String()
		},
		&apps.Deployment{},
		"apps.Deployment",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			return this.clients.Kube().CreateDeployment(namespace, value.(*apps.Deployment))
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.Kube().PatchDeployment(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*apps.Deployment).GetName())
		},
	)
}

func (this *KubePatcher) reloadService() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SERVICE); exists {
		r, e := this.clients.Kube().
			GetService(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_SERVICE)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_SERVICE, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	}
}

func (this *KubePatcher) patchService() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_SERVICE,
		func(value interface{}) string {
			return value.(*core.Service).String()
		},
		&core.Service{},
		"core.Service",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			return this.clients.Kube().CreateService(namespace, value.(*core.Service))
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.Kube().PatchService(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*core.Service).GetName())
		},
	)
}

func (this *KubePatcher) reloadIngress() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_INGRESS); exists {
		r, e := this.clients.Kube().
			GetIngress(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_INGRESS)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_INGRESS, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	}
}

func (this *KubePatcher) patchIngress() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_INGRESS,
		func(value interface{}) string {
			return value.(*extensions.Ingress).String()
		},
		&extensions.Ingress{},
		"extensions.Ingress",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			return this.clients.Kube().CreateIngress(namespace, value.(*extensions.Ingress))
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.Kube().PatchIngress(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*extensions.Ingress).GetName())
		},
	)
}

func (this *KubePatcher) reloadPodDisruptionBudget() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_POD_DISRUPTION_BUDGET); exists {
		r, e := this.clients.Kube().
			GetPodDisruptionBudget(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_POD_DISRUPTION_BUDGET)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_POD_DISRUPTION_BUDGET, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	}
}

func (this *KubePatcher) patchPodDisruptionBudget() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_POD_DISRUPTION_BUDGET,
		func(value interface{}) string {
			return value.(*policy.PodDisruptionBudget).String()
		},
		&policy.PodDisruptionBudget{},
		"policy.PodDisruptionBudget",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			return this.clients.Kube().CreatePodDisruptionBudget(namespace, value.(*policy.PodDisruptionBudget))
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.Kube().PatchPodDisruptionBudget(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*policy.PodDisruptionBudget).GetName())
		},
	)
}

// =====

func (this *KubePatcher) Reload() {
	this.reloadApicurioRegistry()
	this.reloadDeployment()
	this.reloadService()
	this.reloadIngress()
	this.reloadPodDisruptionBudget()
}

func (this *KubePatcher) Execute() {
	this.patchApicurioRegistry()
	this.patchDeployment()
	this.patchService()
	this.patchIngress()
	this.patchPodDisruptionBudget()
}
