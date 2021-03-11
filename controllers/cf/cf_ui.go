package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
)

var _ loop.ControlFunction = &UICF{}

const ENV_UI_READ_ONLY = "REGISTRY_UI_FEATURES_READONLY"

type UICF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	UIReadOnly       bool
	valid            bool
	envUIReadOnly    string
}

func NewUICF(ctx *context.LoopContext) loop.ControlFunction {
	return &UICF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		UIReadOnly:       false,
		valid:            true,
		envUIReadOnly:    "",
	}
}

func (this *UICF) Describe() string {
	return "UICF"
}

func (this *UICF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		this.UIReadOnly = spec.Spec.Configuration.UI.ReadOnly
		// Default value is false
	}

	// Observation #2
	// Read the env values
	this.envUIReadOnly = ""
	if val, exists := this.svcEnvCache.Get(ENV_UI_READ_ONLY); exists {
		this.envUIReadOnly = val.GetValue().Value
	}

	// We won't actively delete old env values if not used
}

func (this *UICF) Compare() bool {
	// Condition #1
	// Has the value changed
	return (this.UIReadOnly == true && this.envUIReadOnly != "true") ||
		(this.UIReadOnly == false && this.envUIReadOnly != "false" && this.envUIReadOnly != "")
}

func (this *UICF) Respond() {
	// Response #1
	// Just set the value(s)!
	val := "false"
	if this.UIReadOnly {
		val = "true"
	}
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_UI_READ_ONLY, val))
}

func (this *UICF) Cleanup() bool {
	// No cleanup
	return true
}
