package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
)

var _ loop.ControlFunction = &HostInitCF{}

type HostInitCF struct {
	ctx        loop.ControlLoopContext
	isFirstRun bool
	targetHost string
	specEntry  resources.ResourceCacheEntry
}

// This CF makes sure number of host is aligned
// If there is some other way of determining the number of host needed outside of CR,
// modify the Sense stage so this CF knows about it
func NewHostInitCF(ctx loop.ControlLoopContext) loop.ControlFunction {
	return &HostInitCF{
		ctx:        ctx,
		isFirstRun: true,
		targetHost: "",
		specEntry:  nil,
	}
}

func (this *HostInitCF) Describe() string {
	return "HostInitCF"
}

func (this *HostInitCF) Sense() {
	// Optimization
	if !this.isFirstRun {
		return
	}

	// Observation #4
	// Get spec for patching & the target host
	if specEntry, exists := this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Get(resources.RC_KEY_SPEC); exists {
		this.specEntry = specEntry
		this.targetHost = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
	}

}

func (this *HostInitCF) Compare() bool {
	// Condition #1
	// First run & no host set
	condition := this.isFirstRun && this.targetHost == ""
	// We are going to try this only once
	this.isFirstRun = false
	return condition
}

func (this *HostInitCF) Respond() {
	// Response #1
	// Patch the resource
	this.specEntry.ApplyPatch(func(value interface{}) interface{} {
		spec := value.(*ar.ApicurioRegistry).DeepCopy()
		dotNamespace := "." + this.ctx.GetAppNamespace()
		if dotNamespace == ".default" {
			dotNamespace = ""
		}
		spec.Spec.Deployment.Host = this.ctx.GetAppName() + dotNamespace
		return spec
	})
}

func (this *HostInitCF) Cleanup() bool {
	// No cleanup
	return true
}
