package kafkasql

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v2"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
)

var _ loop.ControlFunction = &KafkasqlCF{}

const (
	PERSISTENCE_ID              = "kafkasql"
	ENV_KAFKA_BOOTSTRAP_SERVERS = "KAFKA_BOOTSTRAP_SERVERS"
)

type KafkasqlCF struct {
	ctx                 *context.LoopContext
	svcResourceCache    resources.ResourceCache
	svcEnvCache         env.EnvCache
	persistence         string
	bootstrapServers    string
	valid               bool
	envBootstrapServers string
}

func NewKafkasqlCF(ctx *context.LoopContext) loop.ControlFunction {
	return &KafkasqlCF{
		ctx:                 ctx,
		svcResourceCache:    ctx.GetResourceCache(),
		svcEnvCache:         ctx.GetEnvCache(),
		persistence:         "",
		bootstrapServers:    "",
		valid:               true,
		envBootstrapServers: "",
	}
}

func (this *KafkasqlCF) Describe() string {
	return "KafkasqlCF"
}

func (this *KafkasqlCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		this.persistence = spec.Spec.Configuration.Persistence
		this.bootstrapServers = spec.Spec.Configuration.Kafkasql.BootstrapServers
	}

	// Observation #2 + #3
	// Is the correct persistence type selected?
	// Validate the config values
	this.valid = this.persistence == PERSISTENCE_ID && this.bootstrapServers != ""

	// Observation #4
	// Read the env values
	if val, exists := this.svcEnvCache.Get(ENV_KAFKA_BOOTSTRAP_SERVERS); exists {
		this.envBootstrapServers = val.GetValue().Value
	}

	// We won't actively delete old env values if not used
}

func (this *KafkasqlCF) Compare() bool {
	// Condition #1
	// Config values are valid
	// Condition #2 + #3
	// The required env vars are not present OR they differ
	return this.valid && (this.bootstrapServers != this.envBootstrapServers)
}

func (this *KafkasqlCF) Respond() {
	// Response #1
	// Just set the value(s)!
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_KAFKA_BOOTSTRAP_SERVERS, this.bootstrapServers))
}

func (this *KafkasqlCF) Cleanup() bool {
	// No cleanup
	return true
}
