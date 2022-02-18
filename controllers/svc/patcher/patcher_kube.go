package patcher

import (
	goctx "context"
	"errors"
	"reflect"

	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"

	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type KubePatcher struct {
	ctx         *context.LoopContext
	clients     *client.Clients
	factoryKube *factory.KubeFactory
	status      *status.Status
}

func NewKubePatcher(ctx *context.LoopContext, clients *client.Clients, factoryKube *factory.KubeFactory, status *status.Status) *KubePatcher {
	return &KubePatcher{
		ctx:         ctx,
		clients:     clients,
		factoryKube: factoryKube,
		status:      status,
	}
}

// ===

func (this *KubePatcher) reloadApicurioRegistry() {
	// No need to check if the entry exists
	r, e := this.clients.CRD().GetApicurioRegistry(this.ctx.GetAppNamespace(), this.ctx.GetAppName(), &meta.GetOptions{})
	if e != nil {
		this.ctx.GetLog().WithValues("name", this.ctx.GetAppName()).Error(e, "Resource not found. (May have been deleted).")
		this.ctx.GetResourceCache().Remove(resources.RC_KEY_SPEC)
		this.ctx.SetRequeueNow()
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

// MUST be called after reloadApicurioRegistry
func (this *KubePatcher) reloadApicurioRegistryStatus() {
	// No need to check if the entry exists
	specEntry, specExists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SPEC)
	if !specExists {
		this.ctx.GetLog().WithValues("name", this.ctx.GetAppName()).
			Error(errors.New("Could not reload ApicurioRegistryStatus. ApicurioRegistry resource not found."), "Resource not found. (May have been deleted).")
		this.ctx.GetResourceCache().Remove(resources.RC_KEY_SPEC)
		this.ctx.GetResourceCache().Remove(resources.RC_KEY_STATUS)
		this.ctx.SetRequeueNow() // TODO Maybe unnecessary
	} else {
		s := specEntry.GetValue().(*ar.ApicurioRegistry).Status.DeepCopy()
		this.ctx.GetResourceCache().Set(resources.RC_KEY_STATUS,
			resources.NewResourceCacheEntry(common.Name(specEntry.GetName()), s))
	}
}

func (this *KubePatcher) patchApicurioRegistryStatus() {

	specEntry, specExists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SPEC)
	statusEntry, statusExists := this.ctx.GetResourceCache().Get(resources.RC_KEY_STATUS)
	if specExists && statusExists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		targetStatus := statusEntry.GetValue().(*ar.ApicurioRegistryStatus)

		if !reflect.DeepEqual(spec.Status, targetStatus) {
			cr := spec.DeepCopy()
			cr.Status = *targetStatus.DeepCopy()
			err := this.ctx.GetClient().Status().Patch(goctx.TODO(), cr, k8sclient.Merge)
			if err != nil {
				this.ctx.GetLog().WithValues("name", specEntry.GetName()).Error(err, "Resource not found. (May have been deleted).")
				this.ctx.GetResourceCache().Remove(resources.RC_KEY_SPEC)
				this.ctx.SetRequeueNow()
			}
		}
	}
}

func (this *KubePatcher) reloadDeployment() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_DEPLOYMENT); exists {
		r, e := this.clients.Kube().GetDeployment(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Error(e, "Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_DEPLOYMENT)
			this.ctx.SetRequeueNow()
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
			this.ctx.GetLog().WithValues("name", entry.GetName()).Error(e, "Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_SERVICE)
			this.ctx.SetRequeueNow()
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
			this.ctx.GetLog().WithValues("name", entry.GetName()).Error(e, "Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_INGRESS)
			this.ctx.SetRequeueNow()
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
			return value.(*networking.Ingress).String()
		},
		&networking.Ingress{},
		"networking.Ingress",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			return this.clients.Kube().CreateIngress(namespace, value.(*networking.Ingress))
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.Kube().PatchIngress(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*networking.Ingress).GetName())
		},
	)
}

func (this *KubePatcher) reloadNetworkPolicy() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_NETWORK_POLICY); exists {
		r, e := this.clients.Kube().
			GetIngress(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Error(e, "Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_NETWORK_POLICY)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_NETWORK_POLICY, resources.NewResourceCacheEntry(common.Name(r.Name), r))
		}
	}
}

func (this *KubePatcher) patchNetworkPolicy() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_NETWORK_POLICY,
		func(value interface{}) string {
			return value.(*networking.NetworkPolicy).String()
		},
		&networking.NetworkPolicy{},
		"networking.NetworkPolicy",
		func(namespace common.Namespace, value interface{}) (interface{}, error) {
			return this.clients.Kube().CreateNetworkPolicy(namespace, value.(*networking.NetworkPolicy))
		},
		func(namespace common.Namespace, name common.Name, data []byte) (interface{}, error) {
			return this.clients.Kube().PatchNetworkPolicy(namespace, name, data)
		},
		func(value interface{}) common.Name {
			return common.Name(value.(*networking.NetworkPolicy).GetName())
		},
	)
}

func (this *KubePatcher) reloadPodDisruptionBudget() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_POD_DISRUPTION_BUDGET); exists {
		r, e := this.clients.Kube().
			GetPodDisruptionBudget(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().WithValues("name", entry.GetName()).Error(e, "Resource not found. (May have been deleted).")
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_POD_DISRUPTION_BUDGET)
			this.ctx.SetRequeueNow()
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
	this.reloadApicurioRegistryStatus()
	this.reloadDeployment()
	this.reloadService()
	this.reloadIngress()
	this.reloadNetworkPolicy()
	this.reloadPodDisruptionBudget()
}

func (this *KubePatcher) Execute() {
	this.patchApicurioRegistry()
	this.patchApicurioRegistryStatus()
	this.patchDeployment()
	this.patchService()
	this.patchIngress()
	this.patchNetworkPolicy()
	this.patchPodDisruptionBudget()
}
