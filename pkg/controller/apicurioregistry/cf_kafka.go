package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
)

var _ ControlFunction = &KafkaCF{}

const ENV_KAFKA_BOOTSTRAP_SERVERS = "KAFKA_BOOTSTRAP_SERVERS"

type KafkaCF struct {
	ctx                 *Context
	persistence         string
	bootstrapServers    string
	valid               bool
	envBootstrapServers string
}

func NewKafkaCF(ctx *Context) ControlFunction {
	return &KafkaCF{
		ctx:                 ctx,
		persistence:         "",
		bootstrapServers:    "",
		valid:               true,
		envBootstrapServers: "",
	}
}

func (this *KafkaCF) Describe() string {
	return "KafkaCF"
}

func (this *KafkaCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry).Spec
		this.persistence = spec.Configuration.Persistence
		this.bootstrapServers = spec.Configuration.Kafka.BootstrapServers
		// TODO Use secrets!
	}

	// Observation #2 + #3
	// Is the correct persistence type selected?
	// Validate the config values
	this.valid = this.persistence == "kafka" && this.bootstrapServers != ""

	// Observation #4
	// Read the env values
	if val, exists := this.ctx.GetEnvCache().Get(ENV_KAFKA_BOOTSTRAP_SERVERS); exists {
		this.envBootstrapServers = val.GetValue().Value
	}

	// We won't actively delete old env values if not used
}

func (this *KafkaCF) Compare() bool {
	// Condition #1
	// Is JPA & config values are valid
	// Condition #2 + #3
	// The required env vars are not present OR they differ
	return this.valid &&
		(this.bootstrapServers != this.envBootstrapServers)
}

func (this *KafkaCF) Respond() {
	// Response #1
	// Just set the value(s)!
	this.ctx.GetEnvCache().Set(NewSimpleEnvCacheEntry(ENV_KAFKA_BOOTSTRAP_SERVERS, this.bootstrapServers))

}

func (this *KafkaCF) Cleanup() bool {
	// No cleanup
	return true
}
