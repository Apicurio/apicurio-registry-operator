package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
)

var _ loop.ControlFunction = &UpgradeCF{}

type UpgradeCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache

	containerNameUpgradeNeeded bool
	containerNameUpgradeDone   bool
}

func NewUpgradeCF(ctx context.LoopContext) loop.ControlFunction {
	res := &UpgradeCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *UpgradeCF) Describe() string {
	return "UpgradeCF"
}

func (this *UpgradeCF) Sense() {

	// Ensure that the registry container in Deployment is named correctly
	// This will ensure that the previous deployment is properly managed by this operator version
	// Observation #1
	// Get the cached Deployment (if it exists and/or the value)
	this.containerNameUpgradeNeeded = false
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists {
		containers := entry.GetValue().(*apps.Deployment).Spec.Template.Spec.Containers
		oldContainer := common.GetContainerByName(containers, this.ctx.GetAppName().Str())
		newContainer := common.GetContainerByName(containers, factory.REGISTRY_CONTAINER_NAME)
		if oldContainer != nil {
			if newContainer == nil {
				this.containerNameUpgradeNeeded = true
			} else {
				if oldContainer.Name != newContainer.Name {
					this.log.Warnw("cannot upgrade: both containers named " + oldContainer.Name + " and " + newContainer.Name +
						" found in the Deployment")
				} // else, just by coincidence, the CRD is named factory.REGISTRY_CONTAINER_NAME
			}
		}
	}
}

func (this *UpgradeCF) Compare() bool {
	return this.containerNameUpgradeNeeded && !this.containerNameUpgradeDone
}

func (this *UpgradeCF) Respond() {

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists {
		if this.containerNameUpgradeNeeded {
			entry.ApplyPatch(func(value interface{}) interface{} {
				deployment := value.(*apps.Deployment).DeepCopy()
				oldContainer := common.GetContainerByName(deployment.Spec.Template.Spec.Containers, this.ctx.GetAppName().Str())
				oldContainer.Name = factory.REGISTRY_CONTAINER_NAME
				return deployment
			})
			this.containerNameUpgradeDone = true
			this.log.Infow("upgrade successful: renamed container name")
		}
	}
}

func (this *UpgradeCF) Cleanup() bool {
	// No cleanup
	return true
}
