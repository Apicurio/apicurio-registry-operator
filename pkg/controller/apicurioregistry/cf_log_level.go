package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
)

var _ loop.ControlFunction = &LogLevelCF{}

const ENV_REGISTRY_LOG_LEVEL = "LOG_LEVEL"

type LogLevelCF struct {
	ctx         loop.ControlLoopContext
	logLevel    string
	valid       bool
	envLogLevel string
}

func NewLogLevelCF(ctx loop.ControlLoopContext) loop.ControlFunction {
	return &LogLevelCF{
		ctx:         ctx,
		logLevel:    "",
		valid:       true,
		envLogLevel: "",
	}
}

func (this *LogLevelCF) Describe() string {
	return "LogLevelCF"
}

func (this *LogLevelCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(ResourceCache).Get(RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		this.logLevel = spec.Spec.Configuration.LogLevel
		// Default value is false
	}

	// Observation #2
	// Read the env values
	this.envLogLevel = ""
	if val, exists := this.ctx.RequireService(svc.SVC_ENV_CACHE).(EnvCache).Get(ENV_REGISTRY_LOG_LEVEL); exists {
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
	this.ctx.RequireService(svc.SVC_ENV_CACHE).(EnvCache).Set(NewSimpleEnvCacheEntry(ENV_REGISTRY_LOG_LEVEL, this.logLevel))
}

func (this *LogLevelCF) Cleanup() bool {
	// No cleanup
	return true
}
