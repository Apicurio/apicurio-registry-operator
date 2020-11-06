package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	ocp_apps "github.com/openshift/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	policy "k8s.io/api/policy/v1beta1"
)

var _ loop.ControlFunction = &LabelsOcpCF{}

type LabelsOcpCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcOCPFactory    *factory.OCPFactory

	podEntry    resources.ResourceCacheEntry
	podIsCached bool
	podLabels   map[string]string

	caLabels map[string]string

	deploymentEntry     resources.ResourceCacheEntry
	deploymentIsCached  bool
	deploymentLabels    map[string]string
	deploymentPodLabels map[string]string
	updateDeployment    bool
	updateDeploymentPod bool

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
func NewLabelsOcpCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {
	return &LabelsOcpCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcOCPFactory:    services.OcpFactory,
		podLabels:        nil,
	}
}

func (this *LabelsOcpCF) Describe() string {
	return "LabelsOcpCF"
}

func (this *LabelsOcpCF) Sense() {
	// Observation #1
	// Operator Pod
	this.podEntry, this.podIsCached = this.svcResourceCache.Get(resources.RC_KEY_OPERATOR_POD)
	if this.podIsCached {
		this.podLabels = this.podEntry.GetValue().(*core.Pod).Labels
	}
	// Observation #2
	// Deployment & Deployment Pod Template
	this.deploymentEntry, this.deploymentIsCached = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT_OCP)
	if this.deploymentIsCached {
		this.deploymentLabels = this.deploymentEntry.GetValue().(*ocp_apps.DeploymentConfig).Labels
		this.deploymentPodLabels = this.deploymentEntry.GetValue().(*ocp_apps.DeploymentConfig).Spec.Template.Labels
	}
	// Observation #3
	// Service
	this.serviceEntry, this.serviceIsCached = this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
	if this.serviceIsCached {
		this.serviceLabels = this.serviceEntry.GetValue().(*core.Service).Labels
	}
	// Observation #4
	// Ingress
	this.ingressEntry, this.ingressIsCached = this.svcResourceCache.Get(resources.RC_KEY_INGRESS)
	if this.ingressIsCached {
		this.ingressLabels = this.ingressEntry.GetValue().(*extensions.Ingress).Labels
	}
	// Observation #5
	// PodDisruptionBudget
	this.pdbEntry, this.pdbIsCached = this.svcResourceCache.Get(resources.RC_KEY_POD_DISRUPTION_BUDGET)
	if this.pdbIsCached {
		this.pdbLabels = this.pdbEntry.GetValue().(*policy.PodDisruptionBudget).Labels
	}
}

func (this *LabelsOcpCF) Compare() bool {
	this.caLabels = this.GetCommonApplicationLabels()
	this.updateDeployment = this.deploymentIsCached && !LabelsEqual(this.deploymentLabels, this.caLabels)
	this.updateDeploymentPod = this.deploymentIsCached && !LabelsEqual(this.deploymentPodLabels, this.caLabels)
	this.updateService = this.serviceIsCached && !LabelsEqual(this.serviceLabels, this.caLabels)
	this.updateIngress = this.ingressIsCached && !LabelsEqual(this.ingressLabels, this.caLabels)
	this.updatePdb = this.pdbIsCached && !LabelsEqual(this.pdbLabels, this.caLabels)

	return this.podIsCached && (this.updateDeployment ||
		this.updateDeploymentPod ||
		this.updateService ||
		this.updateIngress ||
		this.updatePdb)
}

func (this *LabelsOcpCF) Respond() {
	// Response #1
	// Patch Deployment
	if this.updateDeployment {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*ocp_apps.DeploymentConfig).DeepCopy()
			LabelsUpdate(deployment.Labels, this.caLabels)
			return deployment
		})
	}
	// Response #2
	// Patch Deployment Pod Template
	if this.updateDeploymentPod {
		this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*ocp_apps.DeploymentConfig).DeepCopy()
			LabelsUpdate(deployment.Spec.Template.Labels, this.caLabels)
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
			ingress := value.(*extensions.Ingress).DeepCopy()
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

func (this *LabelsOcpCF) Cleanup() bool {
	// No cleanup
	return true
}

// ---

func (this *LabelsOcpCF) GetCommonApplicationLabels() map[string]string {
	return this.svcOCPFactory.GetLabels()
}
