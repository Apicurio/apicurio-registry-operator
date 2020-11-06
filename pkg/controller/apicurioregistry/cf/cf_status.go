package cf

import (
	"reflect"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/status"
)

var _ loop.ControlFunction = &StatusCF{}

type StatusCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcStatus        *status.Status
	svcKubeFactory   *factory.KubeFactory

	specEntry      resources.ResourceCacheEntry
	specExists     bool
	existingStatus ar.ApicurioRegistryStatus
	targetStatus   ar.ApicurioRegistryStatus
}

// This CF updates the status part of ApicurioRegistry resource
func NewStatusCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {
	return &StatusCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcStatus:        ctx.GetStatus(),
		svcKubeFactory:   services.KubeFactory,
	}
}

func (this *StatusCF) Describe() string {
	return "StatusCF"
}

func (this *StatusCF) Sense() {

	this.specEntry, this.specExists = this.svcResourceCache.Get(resources.RC_KEY_SPEC)
	if this.specExists {
		spec := this.specEntry.GetValue().(*ar.ApicurioRegistry)
		this.existingStatus = spec.Status
		this.targetStatus = *this.svcKubeFactory.CreateStatus(spec)
	}
}

func (this *StatusCF) Compare() bool {
	return this.specExists && !reflect.DeepEqual(this.existingStatus, this.targetStatus)
}

func (this *StatusCF) Respond() {
	this.specEntry.ApplyPatch(func(value interface{}) interface{} {
		spec := value.(*ar.ApicurioRegistry).DeepCopy()
		spec.Status = this.targetStatus
		return spec
	})
}

func (this *StatusCF) Cleanup() bool {
	// No cleanup
	return true
}
