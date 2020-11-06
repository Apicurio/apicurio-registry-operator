package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
)

var _ loop.ControlFunction = &LogLevelCF{}

const ENV_REGISTRY_LOG_LEVEL = "LOG_LEVEL"

type LogLevelCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	logLevel         string
	valid            bool
	envLogLevel      string
}

func NewLogLevelCF(ctx *context.LoopContext) loop.ControlFunction {
	return &LogLevelCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		logLevel:         "",
		valid:            true,
		envLogLevel:      "",
	}
}

func (this *LogLevelCF) Describe() string {
	return "LogLevelCF"
}

func (this *LogLevelCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		this.logLevel = spec.Spec.Configuration.LogLevel
		// Default value is false
	}

	// Observation #2
	// Read the env values
	this.envLogLevel = ""
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_LOG_LEVEL); exists {
		this.envLogLevel = val.GetValue().Value
	}

	// TODO log level validation?

	// We won't actively delete old env values if not used
}

func (this *LogLevelCF) Compare() bool {
	// Condition #1
	// Has the value changed
	return this.logLevel != this.envLogLevel
}

func (this *LogLevelCF) Respond() {
	// Response #1
	// Just set the value(s)!
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_LOG_LEVEL, this.logLevel))
}

func (this *LogLevelCF) Cleanup() bool {
	// No cleanup
	return true
}
