package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy "k8s.io/api/policy/v1beta1"
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

	pdbEntry    resources.ResourceCacheEntry
	pdbIsCached bool
	pdbLabels   map[string]string
	updatePdb   bool
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
	this.pdbEntry, this.pdbIsCached = this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET)
	if this.pdbIsCached {
		this.pdbLabels = this.pdbEntry.GetValue().(*policy.PodDisruptionBudget).Labels
	}
}

func (this *LabelsCF) Compare() bool {
	this.caLabels = this.GetCommonApplicationLabels()
	this.updateDeployment = this.deploymentIsCached && !LabelsEqual(this.deploymentLabels, this.caLabels)
	this.targetDeploymentPodLabels = this.GetTargetDeploymentPodLabels()
	this.updateDeploymentPod = this.deploymentIsCached && !LabelsEqual(this.deploymentPodLabels, this.targetDeploymentPodLabels)
	this.updateService = this.serviceIsCached && !LabelsEqual(this.serviceLabels, this.caLabels)
	this.updateIngress = this.ingressIsCached && !LabelsEqual(this.ingressLabels, this.caLabels)
	this.updatePdb = this.pdbIsCached && !LabelsEqual(this.pdbLabels, this.caLabels)

	return (this.updateDeployment ||
		this.updateDeploymentPod ||
		this.updateService ||
		this.updateIngress ||
		this.updatePdb)
}

func (this *LabelsCF) Respond() {
	// Response #1
	// Patch Deployment
	if this.updateDeployment {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			LabelsUpdate(deployment.Labels, this.caLabels)
			return deployment
		})
	}
	// Response #2
	// Patch Deployment Pod Template
	if this.updateDeploymentPod {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			LabelsUpdate(deployment.Spec.Template.Labels, this.targetDeploymentPodLabels)
			return deployment
		})
	}
	// Response #3
	// Service
	if this.updateService {
		this.serviceEntry.ApplyPatch(func(value interface{}) interface{} {
			service := value.(*core.Service).DeepCopy()
			LabelsUpdate(service.Labels, this.caLabels)
			return service
		})
	}
	// Response #4
	// Ingress
	if this.updateIngress {
		this.ingressEntry.ApplyPatch(func(value interface{}) interface{} {
			ingress := value.(*networking.Ingress).DeepCopy()
			LabelsUpdate(ingress.Labels, this.caLabels)
			return ingress
		})
	}
	// Response #5
	// PodDisruptionBudget
	if this.updatePdb {
		this.pdbEntry.ApplyPatch(func(value interface{}) interface{} {
			pdb := value.(*policy.PodDisruptionBudget).DeepCopy()
			LabelsUpdate(pdb.Labels, this.caLabels)
			return pdb
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
	LabelsUpdate(targetDeploymentPodLabels, this.GetCommonApplicationLabels())
	LabelsUpdate(targetDeploymentPodLabels, this.additionalDeploymentPodLabels)
	return targetDeploymentPodLabels
}

// Return *true* if, for given source labels,
// the target label values exist and have the same value
func LabelsEqual(target map[string]string, source map[string]string) bool {
	for sourceKey, sourceValue := range source {
		targetValue, targetExists := target[sourceKey]
		if !targetExists || sourceValue != targetValue {
			return false
		}
	}
	return true
}

func LabelsUpdate(target map[string]string, source map[string]string) {
	for sourceKey, sourceValue := range source {
		targetValue, targetExists := target[sourceKey]
		if !targetExists || sourceValue != targetValue {
			target[sourceKey] = sourceValue
		}
	}
}
