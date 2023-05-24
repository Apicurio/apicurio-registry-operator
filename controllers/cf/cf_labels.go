package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy_v1 "k8s.io/api/policy/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
)

var _ loop.ControlFunction = &LabelsCF{}

type LabelsCF struct {
	ctx              context.LoopContext
	svcResourceCache resources.ResourceCache
	svcKubeFactory   *factory.KubeFactory

	caLabels map[string]string

	deploymentEntry               resources.ResourceCacheEntry
	deploymentIsCached            bool
	deploymentLabels              map[string]string
	deploymentPodLabels           map[string]string
	additionalDeploymentPodLabels map[string]string
	targetDeploymentPodLabels     map[string]string
	updateDeployment              bool
	updateDeploymentPod           bool

	serviceEntry    resources.ResourceCacheEntry
	serviceIsCached bool
	serviceLabels   map[string]string
	updateService   bool

	ingressEntry    resources.ResourceCacheEntry
	ingressIsCached bool
	ingressLabels   map[string]string
	updateIngress   bool

	pdbV1beta1Entry    resources.ResourceCacheEntry
	pdbV1beta1IsCached bool
	pdbV1beta1Labels   map[string]string
	updatePdbV1beta1   bool

	pdbV1Entry    resources.ResourceCacheEntry
	pdbV1IsCached bool
	pdbV1Labels   map[string]string
	updatePdbV1   bool

	networkPolicyEntry    resources.ResourceCacheEntry
	networkPolicyIsCached bool
	networkPolicyLabels   map[string]string
	updateNetworkPolicy   bool
}

// Update labels on some managed resources
func NewLabelsCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	return &LabelsCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcKubeFactory:   services.GetKubeFactory(),
	}
}

func (this *LabelsCF) Describe() string {
	return "LabelsCF"
}

func (this *LabelsCF) Sense() {
	// Observation #1
	// Deployment & Deployment Pod Template
	this.deploymentEntry, this.deploymentIsCached = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)
	if this.deploymentIsCached {
		this.deploymentLabels = this.deploymentEntry.GetValue().(*apps.Deployment).Labels
		this.deploymentPodLabels = this.deploymentEntry.GetValue().(*apps.Deployment).Spec.Template.Labels
	}
	// Get any additional Deployment Pod labels from the spec
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.additionalDeploymentPodLabels = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Metadata.Labels
	}
	// Observation #2
	// Service
	this.serviceEntry, this.serviceIsCached = this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
	if this.serviceIsCached {
		this.serviceLabels = this.serviceEntry.GetValue().(*core.Service).Labels
	}
	// Observation #3
	// Ingress
	this.ingressEntry, this.ingressIsCached = this.svcResourceCache.Get(resources.RC_KEY_INGRESS)
	if this.ingressIsCached {
		this.ingressLabels = this.ingressEntry.GetValue().(*networking.Ingress).Labels
	}
	// Observation #4
	// PodDisruptionBudget
	this.pdbV1beta1Entry, this.pdbV1beta1IsCached = this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1BETA1)
	if this.pdbV1beta1IsCached {
		this.pdbV1beta1Labels = this.pdbV1beta1Entry.GetValue().(*policy_v1beta1.PodDisruptionBudget).Labels
	}
	this.pdbV1Entry, this.pdbV1IsCached = this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET_V1)
	if this.pdbV1IsCached {
		this.pdbV1Labels = this.pdbV1Entry.GetValue().(*policy_v1.PodDisruptionBudget).Labels
	}
	// Observation #5
	// NetworkPolicy
	this.networkPolicyEntry, this.networkPolicyIsCached = this.svcResourceCache.Get(resources.RC_KEY_NETWORK_POLICY)
	if this.networkPolicyIsCached {
		this.networkPolicyLabels = this.networkPolicyEntry.GetValue().(*networking.NetworkPolicy).Labels
	}
}

func (this *LabelsCF) Compare() bool {
	this.caLabels = this.GetCommonApplicationLabels()
	this.updateDeployment = this.deploymentIsCached && !common.LabelsEqual(this.deploymentLabels, this.caLabels)
	this.targetDeploymentPodLabels = this.GetTargetDeploymentPodLabels()
	this.updateDeploymentPod = this.deploymentIsCached && !common.LabelsEqual(this.deploymentPodLabels, this.targetDeploymentPodLabels)
	this.updateService = this.serviceIsCached && !common.LabelsEqual(this.serviceLabels, this.caLabels)
	this.updateIngress = this.ingressIsCached && !common.LabelsEqual(this.ingressLabels, this.caLabels)
	this.updatePdbV1beta1 = this.pdbV1beta1IsCached && !common.LabelsEqual(this.pdbV1beta1Labels, this.caLabels)
	this.updatePdbV1 = this.pdbV1IsCached && !common.LabelsEqual(this.pdbV1Labels, this.caLabels)
	this.updateNetworkPolicy = this.networkPolicyIsCached && !common.LabelsEqual(this.networkPolicyLabels, this.caLabels)
	// TODO Add network policy labels

	return this.updateDeployment ||
		this.updateDeploymentPod ||
		this.updateService ||
		this.updateIngress ||
		this.updatePdbV1beta1 ||
		this.updatePdbV1 ||
		this.updateNetworkPolicy
}

func (this *LabelsCF) Respond() {
	// Response #1
	// Patch Deployment
	if this.updateDeployment {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			common.LabelsUpdate(deployment.Labels, this.caLabels)
			return deployment
		})
	}
	// Response #2
	// Patch Deployment Pod Template
	if this.updateDeploymentPod {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			common.LabelsUpdate(deployment.Spec.Template.Labels, this.targetDeploymentPodLabels)
			return deployment
		})
	}
	// Response #3
	// Service
	if this.updateService {
		this.serviceEntry.ApplyPatch(func(value interface{}) interface{} {
			service := value.(*core.Service).DeepCopy()
			common.LabelsUpdate(service.Labels, this.caLabels)
			return service
		})
	}
	// Response #4
	// Ingress
	if this.updateIngress {
		this.ingressEntry.ApplyPatch(func(value interface{}) interface{} {
			ingress := value.(*networking.Ingress).DeepCopy()
			common.LabelsUpdate(ingress.Labels, this.caLabels)
			return ingress
		})
	}
	// Response #5
	// PodDisruptionBudget
	if this.updatePdbV1beta1 {
		this.pdbV1beta1Entry.ApplyPatch(func(value interface{}) interface{} {
			pdb := value.(*policy_v1beta1.PodDisruptionBudget).DeepCopy()
			common.LabelsUpdate(pdb.Labels, this.caLabels)
			return pdb
		})
	}
	if this.updatePdbV1 {
		this.pdbV1Entry.ApplyPatch(func(value interface{}) interface{} {
			pdb := value.(*policy_v1.PodDisruptionBudget).DeepCopy()
			common.LabelsUpdate(pdb.Labels, this.caLabels)
			return pdb
		})
	}
	// Response #4
	// NetworkPolicy
	if this.updateNetworkPolicy {
		this.networkPolicyEntry.ApplyPatch(func(value interface{}) interface{} {
			policy := value.(*networking.NetworkPolicy).DeepCopy()
			common.LabelsUpdate(policy.Labels, this.caLabels)
			return policy
		})
	}
}

func (this *LabelsCF) Cleanup() bool {
	// No cleanup
	return true
}

// ---

func (this *LabelsCF) GetCommonApplicationLabels() map[string]string {
	return this.svcKubeFactory.GetLabels()
}

func (this *LabelsCF) GetTargetDeploymentPodLabels() map[string]string {
	targetDeploymentPodLabels := make(map[string]string)
	common.LabelsUpdate(targetDeploymentPodLabels, this.GetCommonApplicationLabels())
	common.LabelsUpdate(targetDeploymentPodLabels, this.additionalDeploymentPodLabels)
	return targetDeploymentPodLabels
}
