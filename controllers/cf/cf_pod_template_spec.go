package cf

import (
	"encoding/json"
	"errors"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	f "github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"reflect"
)

var _ loop.ControlFunction = &PodTemplateSpecCF{}

type PodTemplateSpecCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache
	services         services.LoopServices

	previousBasePodTemplateSpec *ar.ApicurioRegistryPodTemplateSpec
	basePodTemplateSpec         *ar.ApicurioRegistryPodTemplateSpec

	previousDeploymentPodSpec *core.PodTemplateSpec
	deploymentPodSpec         *core.PodTemplateSpec

	valid                 bool
	targetPodTemplateSpec *core.PodTemplateSpec

	lastActedReconcileSequence int64
}

// Using the Deployment directly to detemine whether the PTS has to be applied is a problem,
// because other CFs might modify the Deployment as well.
// We have no way of comparing the target PTS with the PTS in the Deployment,
// since we don't know which changes are from spec PTS and which from the CFs.
// To work around this, we will reconcile if:
// - The PTS in the spec has changed, or
// - The PTS in the Deployment has changed.
// If we are done, there will be no change in the Deployment between execution of this CF in subsequent reconciliations,
// so we don't have to update the PTS again. This may waste a few cycles, but I don't think we can do better than that.
func NewPodTemplateSpecCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &PodTemplateSpecCF{
		ctx:                        ctx,
		svcResourceCache:           ctx.GetResourceCache(),
		services:                   services,
		lastActedReconcileSequence: -2,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *PodTemplateSpecCF) Describe() string {
	return "PodTemplateSpecCF"
}

func (this *PodTemplateSpecCF) Sense() {

	this.log.Debugw("Sense",
		"this.ctx.GetReconcileSequence()", this.ctx.GetReconcileSequence(),
		"this.lastActedReconcileSequence", this.lastActedReconcileSequence,
	)

	if this.lastActedReconcileSequence+1 == this.ctx.GetReconcileSequence() {
		this.log.Debugln("Sense", "We have acted in the previous loop, record the previous PTS from the Deployment, and reschedule")
		// We have acted in the previous loop, record the previous PTS from the Deployment, and reschedule
		if deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); deploymentExists {
			this.log.Debugln("Sense", "Setting this.previousDeploymentPodSpec")
			this.previousDeploymentPodSpec = &deploymentEntry.GetValue().(*apps.Deployment).Spec.Template
			this.previousDeploymentPodSpec = this.previousDeploymentPodSpec.DeepCopy() // Defensive copy
			this.ctx.SetRequeueNow()
			return
		}
	}

	this.valid = false

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {

		this.basePodTemplateSpec = &entry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.PodTemplateSpecPreview
		this.basePodTemplateSpec = this.basePodTemplateSpec.DeepCopy() // Defensive copy so we won't update the spec

		if deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); deploymentExists {
			this.deploymentPodSpec = &deploymentEntry.GetValue().(*apps.Deployment).Spec.Template
			this.deploymentPodSpec = this.deploymentPodSpec.DeepCopy() // Defensive copy
			factoryPodSpec := this.services.GetKubeFactory().CreateDeployment().Spec.Template
			targetPodSpec, err := SanitizeBasePodSpec(this.log, this.basePodTemplateSpec, this.deploymentPodSpec, &factoryPodSpec)
			if err == nil {
				this.targetPodTemplateSpec, err = ConvertToPodTemplateSpec(targetPodSpec)
				if err == nil {
					this.valid = true
				} else {
					this.log.Errorw("an error has occurred when processing spec.deployment.podTemplateSpecPreview field", "error", err)
				}
			} else {
				this.log.Errorw("an error has occurred when processing spec.deployment.podTemplateSpecPreview field", "error", err)
				this.services.GetConditionManager().GetConfigurationErrorCondition().
					TransitionInvalid(err.Error(), "spec.deployment.podTemplateSpecPreview")
				// No need to transition to not-ready
				this.ctx.SetRequeueDelaySec(10)
			}
		}
	}
}

func (this *PodTemplateSpecCF) Compare() bool {

	if this.lastActedReconcileSequence == this.ctx.GetReconcileSequence() {
		this.log.Debugln("Compare", "Act only once per reconciliation loop")
		return false // Act only once per reconciliation loop
	}
	if this.lastActedReconcileSequence+1 == this.ctx.GetReconcileSequence() {
		this.log.Debugln("Compare", "We have acted in the previous loop, recorded the previous PTS from the Deployment, and have to skip")
		return false // We have acted in the reconciliation, recorded the previous PTS from the Deployment, and have to skip
	}

	this.log.Debugw("Compare", "valid", this.valid)

	if this.previousBasePodTemplateSpec != nil { // common.LogNillable does not work TODO find out why
		this.log.Debugw("Compare", "this.previousBasePodTemplateSpec", this.previousBasePodTemplateSpec)
	} else {
		this.log.Debugw("Compare", "this.previousBasePodTemplateSpec", "<nil>")
	}
	this.log.Debugw("Compare", "this.basePodTemplateSpec", this.basePodTemplateSpec)

	if this.previousDeploymentPodSpec != nil { // common.LogNillable does not work TODO find out why
		this.log.Debugw("Compare", "this.previousDeploymentPodSpec", this.previousDeploymentPodSpec)
	} else {
		this.log.Debugw("Compare", "this.previousDeploymentPodSpec", "<nil>")
	}
	this.log.Debugw("Compare", "this.deploymentPodSpec", this.deploymentPodSpec)

	this.log.Debugw("Compare", "!reflect.DeepEqual(this.basePodTemplateSpec, this.previousBasePodTemplateSpec)", !reflect.DeepEqual(this.basePodTemplateSpec, this.previousBasePodTemplateSpec))

	this.log.Debugw("Compare", "!reflect.DeepEqual(this.deploymentPodSpec, this.previousDeploymentPodSpec)", !reflect.DeepEqual(this.deploymentPodSpec, this.previousDeploymentPodSpec))

	return this.valid &&
		(!reflect.DeepEqual(this.basePodTemplateSpec, this.previousBasePodTemplateSpec) || !reflect.DeepEqual(this.deploymentPodSpec, this.previousDeploymentPodSpec))
}

func (this *PodTemplateSpecCF) Respond() {
	this.log.Debugln("Respond", "start respond")
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			this.previousBasePodTemplateSpec = this.basePodTemplateSpec
			deployment.Spec.Template = *this.targetPodTemplateSpec
			this.lastActedReconcileSequence = this.ctx.GetReconcileSequence()
			this.log.Debugln("Respond", "responded")
			return deployment
		})
	}
}

func (this *PodTemplateSpecCF) Cleanup() bool {
	// No cleanup
	return true
}

/*
Reserved:

- metadata
  - metadata.annotations [alternative exists]
  - metadata.labels [alternative exists]

- spec
  - spec.affinity [alternative exists]
  - spec.containers[*]
  - spec.containers[name = "registry"].env [alternative exists]
  - spec.containers[name = "registry"].image [alternative exists]
  - spec.imagePullSecrets [alternative exists]
  - spec.tolerations [alternative exists]
*/
func SanitizeBasePodSpec(log *zap.SugaredLogger, base *ar.ApicurioRegistryPodTemplateSpec, current *core.PodTemplateSpec,
	factory *core.PodTemplateSpec) (*ar.ApicurioRegistryPodTemplateSpec, error) {
	// We are using values from *current* with fields (values):
	//   - that the user cannot change
	//   - or are empty by default
	//   - or there is a CF that handles them
	// We are using values from *factory* with fields (values):
	//   - that are not empty by default, so empty fields in the base podSpec do not remove default values
	//   - and there is not and existing CF that handles them
	// Otherwise we just pass the base values along, relying on CFs to override things

	base = base.DeepCopy()       // Defensive copy TODO is this needed?
	current = current.DeepCopy() // Defensive copy TODO is this needed?

	// metadata.annotations
	if base.Metadata.Annotations != nil {
		return nil, newReservedFieldError("metadata.annotations")
	}
	base.Metadata.Annotations = current.ObjectMeta.Annotations

	// metadata.labels
	if base.Metadata.Labels != nil {
		return nil, newReservedFieldError("metadata.labels")
	}
	base.Metadata.Labels = current.ObjectMeta.Labels

	// spec.affinity
	if base.Spec.Affinity != nil {
		return nil, newReservedFieldError("spec.affinity")
	}
	base.Spec.Affinity = current.Spec.Affinity

	// spec.containers[*]

	baseContainer := common.GetContainerByName(base.Spec.Containers, f.REGISTRY_CONTAINER_NAME)
	if len(base.Spec.Containers) > 0 && baseContainer == nil {
		log.Warnw("container named " + f.REGISTRY_CONTAINER_NAME + " is not defined " +
			"in spec.deployment.podTemplateSpecPreview.spec.containers. Make sure this is intended.")
	}
	currentContainer := common.GetContainerByName(current.Spec.Containers, f.REGISTRY_CONTAINER_NAME)
	if currentContainer == nil {
		panic("Could not find the registry container in the current Deployment.")
	}
	factoryContainer := common.GetContainerByName(factory.Spec.Containers, f.REGISTRY_CONTAINER_NAME)
	if factoryContainer == nil {
		panic("Could not find the registry container in the initial Deployment.")
	}

	if len(base.Spec.Containers) > 0 && baseContainer != nil {

		// spec.containers[name = "registry"].env
		if baseContainer.Env != nil || len(baseContainer.Env) != 0 {
			return nil, newReservedFieldError("spec.containers[name = \"registry\"].env")
		}
		baseContainer.Env = currentContainer.Env

		// spec.containers[name = "registry"].image
		if baseContainer.Image != "" {
			return nil, newReservedFieldError("spec.containers[name = \"registry\"].image")
		}
		baseContainer.Image = currentContainer.Image

		// (Factory) spec.containers[name = "registry"].livenessProbe
		if baseContainer.LivenessProbe == nil {
			baseContainer.LivenessProbe = factoryContainer.LivenessProbe
		}

		// (Factory) spec.containers[name = "registry"].readinessProbe
		if baseContainer.ReadinessProbe == nil {
			baseContainer.ReadinessProbe = factoryContainer.ReadinessProbe
		}

		// (Factory) spec.containers[name = "registry"].resources.limits
		if len(baseContainer.Resources.Limits) == 0 {
			baseContainer.Resources.Limits = factoryContainer.Resources.Limits
		}

		// (Factory) spec.containers[name = "registry"].resources.requests
		if len(baseContainer.Resources.Requests) == 0 {
			baseContainer.Resources.Requests = factoryContainer.Resources.Requests
		}

		// (Factory) spec.volumeMounts
		for _, v := range factoryContainer.VolumeMounts {
			common.SetVolumeMount(&baseContainer.VolumeMounts, &v)
		}

	} else {
		base.Spec.Containers = make([]core.Container, len(factory.Spec.Containers))
		for i, c := range factory.Spec.Containers {
			base.Spec.Containers[i] = c
		}
	}

	// spec.imagePullSecrets
	if len(base.Spec.ImagePullSecrets) > 0 {
		return nil, newReservedFieldError("spec.imagePullSecrets")
	}
	base.Spec.ImagePullSecrets = current.Spec.ImagePullSecrets

	// (Factory) spec.terminationGracePeriodSeconds
	if base.Spec.TerminationGracePeriodSeconds == nil {
		base.Spec.TerminationGracePeriodSeconds = factory.Spec.TerminationGracePeriodSeconds
	}

	// spec.tolerations
	if len(base.Spec.Tolerations) > 0 {
		return nil, newReservedFieldError("spec.tolerations")
	}
	base.Spec.Tolerations = current.Spec.Tolerations

	// (Factory) spec.volumes
	for _, v := range factory.Spec.Volumes {
		common.SetVolume(&base.Spec.Volumes, &v)
	}

	return base, nil
}

func newReservedFieldError(field string) error {
	return errors.New("field " + field + " is reserved and must not be defined")
}

// Do a magic using JSON to conver these values.
// They MUST have the equivalent field names.
// We are only doing this because we don't have some ommitempty tags in PodSpec.
func ConvertToPodTemplateSpec(source *ar.ApicurioRegistryPodTemplateSpec) (*core.PodTemplateSpec, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, errors.New("failed to convert between ApicurioRegistryPodTemplateSpec and PodTemplateSpec: " + err.Error())
	}
	out := &core.PodTemplateSpec{}
	err = json.Unmarshal(data, out)
	if err != nil {
		return nil, errors.New("failed to convert between ApicurioRegistryPodTemplateSpec and PodTemplateSpec: " + err.Error())
	}
	return out, nil
}
