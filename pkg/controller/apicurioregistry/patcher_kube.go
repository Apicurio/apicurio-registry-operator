package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubePatcher struct {
	ctx *Context
}

func NewKubePatcher(ctx *Context) *KubePatcher {
	return &KubePatcher{
		ctx: ctx,
	}
}

// ===

func (this *KubePatcher) patchApicurioRegistry() { // TODO move to separate file/class?
	patchGeneric(
		this.ctx,
		RC_KEY_SPEC,
		func(value interface{}) string {
			return value.(*ar.ApicurioRegistry).ObjectMeta.String()
		},
		&ar.ApicurioRegistry{},
		"ar.ApicurioRegistry",
		func(namespace string, value interface{}) (interface{}, error) {
			// This should be not used (at the moment)
			panic("Unsupported operation.")
		},
		func(namespace string, name string, data []byte) (interface{}, error) {
			return this.ctx.GetClients().CRD().PatchApicurioRegistry(namespace, name, data)
		},
		func(value interface{}) string {
			return value.(*ar.ApicurioRegistry).GetName()
		},
	)
}

func (this *KubePatcher) reloadDeployment() {
	if entry, exists := this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT); exists {
		r, e := this.ctx.GetClients().Kube().
			GetDeployment(this.ctx.GetConfiguration().GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(RC_KEY_DEPLOYMENT)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(RC_KEY_DEPLOYMENT, NewResourceCacheEntry(r.Name, r))
		}
	}
}

func (this *KubePatcher) patchDeployment() {
	patchGeneric(
		this.ctx,
		RC_KEY_DEPLOYMENT,
		func(value interface{}) string {
			return value.(*apps.Deployment).String()
		},
		&apps.Deployment{},
		"apps.Deployment",
		func(namespace string, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreateDeployment(namespace, value.(*apps.Deployment))
		},
		func(namespace string, name string, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchDeployment(namespace, name, data)
		},
		func(value interface{}) string {
			return value.(*apps.Deployment).GetName()
		},
	)
}

func (this *KubePatcher) reloadService() {
	if entry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SERVICE); exists {
		r, e := this.ctx.GetClients().Kube().
			GetService(this.ctx.GetConfiguration().GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(RC_KEY_SERVICE)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(RC_KEY_SERVICE, NewResourceCacheEntry(r.Name, r))
		}
	}
}

func (this *KubePatcher) patchService() {
	patchGeneric(
		this.ctx,
		RC_KEY_SERVICE,
		func(value interface{}) string {
			return value.(*core.Service).String()
		},
		&core.Service{},
		"core.Service",
		func(namespace string, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreateService(namespace, value.(*core.Service))
		},
		func(namespace string, name string, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchService(namespace, name, data)
		},
		func(value interface{}) string {
			return value.(*core.Service).GetName()
		},
	)
}

func (this *KubePatcher) reloadIngress() {
	if entry, exists := this.ctx.GetResourceCache().Get(RC_KEY_INGRESS); exists {
		r, e := this.ctx.GetClients().Kube().
			GetIngress(this.ctx.GetConfiguration().GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(RC_KEY_INGRESS)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(RC_KEY_INGRESS, NewResourceCacheEntry(r.Name, r))
		}
	}
}

func (this *KubePatcher) patchIngress() {
	patchGeneric(
		this.ctx,
		RC_KEY_INGRESS,
		func(value interface{}) string {
			return value.(*extensions.Ingress).String()
		},
		&extensions.Ingress{},
		"extensions.Ingress",
		func(namespace string, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreateIngress(namespace, value.(*extensions.Ingress))
		},
		func(namespace string, name string, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchIngress(namespace, name, data)
		},
		func(value interface{}) string {
			return value.(*extensions.Ingress).GetName()
		},
	)
}

func (this *KubePatcher) reloadPodDisruptionBudget() {
	if entry, exists := this.ctx.GetResourceCache().Get(RC_KEY_POD_DISRUPTION_BUDGET); exists {
		r, e := this.ctx.GetClients().Kube().
			GetPodDisruptionBudget(this.ctx.GetConfiguration().GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Info("Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(RC_KEY_POD_DISRUPTION_BUDGET)
			this.ctx.SetRequeue()
		} else {
			this.ctx.GetResourceCache().Set(RC_KEY_POD_DISRUPTION_BUDGET, NewResourceCacheEntry(r.Name, r))
		}
	}
}

func (this *KubePatcher) patchPodDisruptionBudget() {
	patchGeneric(
		this.ctx,
		RC_KEY_POD_DISRUPTION_BUDGET,
		func(value interface{}) string {
			return value.(*policy.PodDisruptionBudget).String()
		},
		&policy.PodDisruptionBudget{},
		"policy.PodDisruptionBudget",
		func(namespace string, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreatePodDisruptionBudget(namespace, value.(*policy.PodDisruptionBudget))
		},
		func(namespace string, name string, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchPodDisruptionBudget(namespace, name, data)
		},
		func(value interface{}) string {
			return value.(*policy.PodDisruptionBudget).GetName()
		},
	)
}

// =====

func (this *KubePatcher) Reload() {
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
