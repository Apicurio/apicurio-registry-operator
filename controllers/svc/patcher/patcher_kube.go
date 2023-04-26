package patcher

import (
	"errors"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"

	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy_v1 "k8s.io/api/policy/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubePatcher struct {
	ctx         context.LoopContext
	factoryKube *factory.KubeFactory
	status      *status.Status
}

func NewKubePatcher(ctx context.LoopContext, factoryKube *factory.KubeFactory, status *status.Status) *KubePatcher {
	return &KubePatcher{
		ctx,
		factoryKube,
		status,
	}
}

// ===

func (this *KubePatcher) reloadApicurioRegistry() {
	//if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SPEC); exists {
	//	r, e := this.ctx.GetClients().CRD().GetApicurioRegistry(this.ctx.GetAppNamespace(), this.ctx.GetAppName())
	//	if e != nil {
	//		this.ctx.GetLog().WithValues("name", entry.GetName()).Error(e, "Resource not found. (May have been deleted).")
	//		this.ctx.GetResourceCache().Remove(resources.RC_KEY_SPEC)
	//		this.ctx.SetRequeueNow()
	//	} else {
	//		this.ctx.GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(c.Name(r.Name), r))
	//	}
	//}
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
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			// This should be not used (at the moment)
			panic("Unsupported operation.")
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().CRD().PatchApicurioRegistry(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return c.Name(value.(*ar.ApicurioRegistry).GetName())
		},
	)
}

// MUST be called after reloadApicurioRegistry
func (this *KubePatcher) reloadApicurioRegistryStatus() {
	// No need to check if the entry exists
	specEntry, specExists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SPEC)
	if !specExists {
		this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
			"name", this.ctx.GetAppName(),
			"error", errors.New("Could not reload ApicurioRegistryStatus. ApicurioRegistry resource not found."))
		this.ctx.GetResourceCache().Remove(resources.RC_KEY_SPEC)
		this.ctx.GetResourceCache().Remove(resources.RC_KEY_STATUS)
		this.ctx.SetRequeueNow() // TODO Maybe unnecessary
	} else {
		s := specEntry.GetValue().(*ar.ApicurioRegistry).Status.DeepCopy()
		this.ctx.GetResourceCache().Set(resources.RC_KEY_STATUS,
			resources.NewResourceCacheEntry(specEntry.GetName(), s))
	}
}

func (this *KubePatcher) patchApicurioRegistryStatus() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_STATUS,
		func(value interface{}) string {
			return "TODO"
		},
		&ar.ApicurioRegistryStatus{},
		"ar.ApicurioRegistryStatus",
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			// This should be not used (at the moment)
			panic("Unsupported operation.")
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().CRD().PatchApicurioRegistryStatus(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return "TODO"
		},
	)
}

/*
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
			resources.NewResourceCacheEntry(c.Name(specEntry.GetName()), s))
	}
}
*/

/*
func (this *KubePatcher) patchApicurioRegistryStatus() {

	specEntry, specExists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SPEC)
	statusEntry, statusExists := this.ctx.GetResourceCache().Get(resources.RC_KEY_STATUS)
	if specExists && statusExists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		targetStatus := statusEntry.GetValue().(*ar.ApicurioRegistryStatus)

		if !reflect.DeepEqual(spec.Status, targetStatus) {
			cr := spec.DeepCopy()
			cr.Status = *targetStatus.DeepCopy()
			err := this.ctx.GetClients().CRD().Status().Patch(goctx.TODO(), cr, k8sclient.Merge)
			if err != nil {
				this.ctx.GetLog().WithValues("name", specEntry.GetName()).Error(err, "Resource not found. (May have been deleted).")
				this.ctx.GetResourceCache().Remove(resources.RC_KEY_SPEC)
				this.ctx.SetRequeueNow()
			}
		}
	}
}
*/

func (this *KubePatcher) reloadDeployment() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_DEPLOYMENT); exists {
		r, e := this.ctx.GetClients().Kube().GetDeployment(this.ctx.GetAppNamespace(), entry.GetName())
		if e != nil {
			this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
				"name", entry.GetName(), "error", e)
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_DEPLOYMENT)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_DEPLOYMENT, resources.NewResourceCacheEntry(c.Name(r.Name), r))
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
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreateDeployment(owner, namespace, value.(*apps.Deployment))
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchDeployment(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return c.Name(value.(*apps.Deployment).GetName())
		},
	)
}

func (this *KubePatcher) reloadService() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_SERVICE); exists {
		r, e := this.ctx.GetClients().Kube().
			GetService(this.ctx.GetAppNamespace(), entry.GetName())
		if e != nil {
			this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
				"name", entry.GetName(), "error", e)
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_SERVICE)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_SERVICE, resources.NewResourceCacheEntry(c.Name(r.Name), r))
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
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreateService(owner, namespace, value.(*core.Service))
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchService(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return c.Name(value.(*core.Service).GetName())
		},
	)
}

func (this *KubePatcher) reloadIngress() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_INGRESS); exists {
		r, e := this.ctx.GetClients().Kube().
			GetIngress(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
				"name", entry.GetName(), "error", e)
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_INGRESS)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_INGRESS, resources.NewResourceCacheEntry(c.Name(r.Name), r))
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
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreateIngress(owner, namespace, value.(*networking.Ingress))
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchIngress(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return c.Name(value.(*networking.Ingress).GetName())
		},
	)
}

func (this *KubePatcher) reloadNetworkPolicy() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_NETWORK_POLICY); exists {
		r, e := this.ctx.GetClients().Kube().
			GetNetworkPolicy(this.ctx.GetAppNamespace(), entry.GetName(), &meta.GetOptions{})
		if e != nil {
			this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
				"name", entry.GetName(), "error", e)
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_NETWORK_POLICY)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_NETWORK_POLICY, resources.NewResourceCacheEntry(c.Name(r.Name), r))
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
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreateNetworkPolicy(owner, namespace, value.(*networking.NetworkPolicy))
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchNetworkPolicy(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return c.Name(value.(*networking.NetworkPolicy).GetName())
		},
	)
}

func (this *KubePatcher) reloadPodDisruptionBudgetV1beta1() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1); exists {
		r, e := this.ctx.GetClients().Kube().
			GetPodDisruptionBudgetV1beta1(this.ctx.GetAppNamespace(), entry.GetName())
		if e != nil {
			this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
				"name", entry.GetName(), "error", e)
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1, resources.NewResourceCacheEntry(c.Name(r.Name), r))
		}
	}
}

func (this *KubePatcher) patchPodDisruptionBudgetV1beta1() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1,
		func(value interface{}) string {
			return value.(*policy_v1beta1.PodDisruptionBudget).String()
		},
		&policy_v1beta1.PodDisruptionBudget{},
		"policy.PodDisruptionBudget",
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreatePodDisruptionBudgetV1beta1(owner, namespace, value.(*policy_v1beta1.PodDisruptionBudget))
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchPodDisruptionBudgetV1beta1(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return c.Name(value.(*policy_v1beta1.PodDisruptionBudget).GetName())
		},
	)
}

func (this *KubePatcher) reloadPodDisruptionBudgetV1() {
	if entry, exists := this.ctx.GetResourceCache().Get(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1); exists {
		r, e := this.ctx.GetClients().Kube().
			GetPodDisruptionBudgetV1(this.ctx.GetAppNamespace(), entry.GetName())
		if e != nil {
			this.ctx.GetLog().Sugar().Warnw("Resource not found. (May have been deleted).",
				"name", entry.GetName(), "error", e)
			this.ctx.GetResourceCache().Remove(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1)
			this.ctx.SetRequeueNow()
		} else {
			this.ctx.GetResourceCache().Set(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1, resources.NewResourceCacheEntry(c.Name(r.Name), r))
		}
	}
}

func (this *KubePatcher) patchPodDisruptionBudgetV1() {
	patchGeneric(
		this.ctx,
		resources.RC_KEY_POD_DISRUPTION_BUDGET_V1,
		func(value interface{}) string {
			return value.(*policy_v1.PodDisruptionBudget).String()
		},
		&policy_v1.PodDisruptionBudget{},
		"policy.PodDisruptionBudget",
		func(owner meta.Object, namespace c.Namespace, value interface{}) (interface{}, error) {
			return this.ctx.GetClients().Kube().CreatePodDisruptionBudgetV1(owner, namespace, value.(*policy_v1.PodDisruptionBudget))
		},
		func(namespace c.Namespace, name c.Name, data []byte) (interface{}, error) {
			return this.ctx.GetClients().Kube().PatchPodDisruptionBudgetV1(namespace, name, data)
		},
		func(value interface{}) c.Name {
			return c.Name(value.(*policy_v1.PodDisruptionBudget).GetName())
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
	this.reloadPodDisruptionBudgetV1beta1()
	this.reloadPodDisruptionBudgetV1()
}

func (this *KubePatcher) Execute() {
	this.patchApicurioRegistry()
	this.patchApicurioRegistryStatus()
	this.patchDeployment()
	this.patchService()
	this.patchIngress()
	this.patchNetworkPolicy()
	this.patchPodDisruptionBudgetV1beta1()
	this.patchPodDisruptionBudgetV1()
}
