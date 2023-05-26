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
	valid                       bool
	targetPodTemplateSpec       *core.PodTemplateSpec
}

func NewPodTemplateSpecCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &PodTemplateSpecCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		services:         services,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *PodTemplateSpecCF) Describe() string {
	return "PodTemplateSpecCF"
}

func (this *PodTemplateSpecCF) Sense() {
	this.valid = false

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {

		this.basePodTemplateSpec = &entry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.PodTemplateSpecPreview
		this.basePodTemplateSpec = this.basePodTemplateSpec.DeepCopy() // Defensive copy so we don't update the spec

		if deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); deploymentExists {
			currentPodSpec := &deploymentEntry.GetValue().(*apps.Deployment).Spec.Template
			currentPodSpec = currentPodSpec.DeepCopy()
			factoryPodSpec := this.services.GetKubeFactory().CreateDeployment().Spec.Template
			targetPodSpec, err := SanitizeBasePodSpec(this.log, this.basePodTemplateSpec, currentPodSpec, &factoryPodSpec)
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
				// No need to transition to not ready, since we can just with the previous config
				this.ctx.SetRequeueDelaySec(10)
			}
		}
	}
}

func (this *PodTemplateSpecCF) Compare() bool {
	if this.previousBasePodTemplateSpec != nil { // common.LogNillable does not work TODO find out why
		this.log.Debugw("Obsevation #1", "this.previousBasePodTemplateSpec", this.previousBasePodTemplateSpec)
	} else {
		this.log.Debugw("Obsevation #1", "this.previousBasePodTemplateSpec", "<nil>")
	}
	this.log.Debugw("Obsevation #2", "this.basePodTemplateSpec", this.basePodTemplateSpec)
	this.log.Debugw("Obsevation #3", "this.targetPodTemplateSpec", this.targetPodTemplateSpec)
	return this.valid &&
		// We're only comparing changes to the podSpecPreview, not the real pod spec,
		// so we do not overwrite changes by the other CFs, which would cause a loop panic
		(this.previousBasePodTemplateSpec == nil || !reflect.DeepEqual(this.basePodTemplateSpec, this.previousBasePodTemplateSpec))
}

func (this *PodTemplateSpecCF) Respond() {

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()

			deployment.Spec.Template = *this.targetPodTemplateSpec

			this.previousBasePodTemplateSpec = this.basePodTemplateSpec

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
    - spec.containers[name = "registry"].imagePullPolicy [alternative exists]
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
		panic("could not find current registry container")
	}
	factoryContainer := common.GetContainerByName(factory.Spec.Containers, f.REGISTRY_CONTAINER_NAME)
	if factoryContainer == nil {
		panic("could not find factory registry container")
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

		// spec.containers[name = "registry"].imagePullPolicy
		if baseContainer.ImagePullPolicy != "" {
			return nil, newReservedFieldError("spec.containers[name = \"registry\"].imagePullPolicy")
		}
		baseContainer.ImagePullPolicy = currentContainer.ImagePullPolicy

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
