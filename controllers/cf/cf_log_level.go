package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
)

var _ loop.ControlFunction = &LogLevelCF{}

const ENV_REGISTRY_LOG_LEVEL = "LOG_LEVEL"
const ENV_REGISTRY_LOG_LEVEL2 = "REGISTRY_LOG_LEVEL"

type LogLevelCF struct {
	ctx              context.LoopContext
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	valid            bool
	logLevel         string
	registryLogLevel string
	envLogLevel      string
	envLogLevel2     string
}

func NewLogLevelCF(ctx context.LoopContext) loop.ControlFunction {
	return &LogLevelCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		valid:            true,
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
		this.registryLogLevel = spec.Spec.Configuration.RegistryLogLevel
		// Default values are false
	}

	// Observation #2
	// Read the env values
	this.envLogLevel = ""
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_LOG_LEVEL); exists {
		this.envLogLevel = val.GetValue().Value
	}
	this.envLogLevel2 = ""
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_LOG_LEVEL2); exists {
		this.envLogLevel2 = val.GetValue().Value
	}

	// TODO log level validation?

	// We won't actively delete old env values if not used
}

func (this *LogLevelCF) Compare() bool {
	// Condition #1
	// Has the value changed
	return this.logLevel != this.envLogLevel || this.registryLogLevel != this.envLogLevel2
}

func (this *LogLevelCF) Respond() {
	// Response #1
	// Just set the value(s)!
	if this.logLevel != "" {
		this.svcEnvCache.Set(env.NewSimpleEnvCacheEntryBuilder(ENV_REGISTRY_LOG_LEVEL, this.logLevel).Build())
	}
	if this.registryLogLevel != "" {
		this.svcEnvCache.Set(env.NewSimpleEnvCacheEntryBuilder(ENV_REGISTRY_LOG_LEVEL2, this.registryLogLevel).Build())
	}
}

func (this *LogLevelCF) Cleanup() bool {
	// No cleanup
	return true
}
