package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	core "k8s.io/api/core/v1"
)

var _ loop.ControlFunction = &StreamsCF{}

// +ENV_KAFKA_BOOTSTRAP_SERVERS
const ENV_APPLICATION_SERVER_HOST = "APPLICATION_SERVER_HOST"
const ENV_APPLICATION_SERVER_PORT = "APPLICATION_SERVER_PORT"
const ENV_APPLICATION_ID = "APPLICATION_ID"

type StreamsCF struct {
	ctx                            *context.LoopContext
	svcResourceCache               resources.ResourceCache
	svcEnvCache                    env.EnvCache
	persistence                    string
	bootstrapServers               string
	applicationServerPort          string
	applicationId                  string
	valid                          bool
	envBootstrapServers            string
	envApplicationServerHostExists bool
	envApplicationServerPort       string
	envApplicationId               string
}

func NewStreamsCF(ctx *context.LoopContext) loop.ControlFunction {
	return &StreamsCF{
		ctx:                            ctx,
		svcResourceCache:               ctx.GetResourceCache(),
		svcEnvCache:                    ctx.GetEnvCache(),
		persistence:                    "",
		bootstrapServers:               "",
		applicationServerPort:          "",
		applicationId:                  "",
		valid:                          true,
		envBootstrapServers:            "",
		envApplicationServerHostExists: false,
		envApplicationServerPort:       "",
		envApplicationId:               "",
	}
}

func (this *StreamsCF) Describe() string {
	return "StreamsCF"
}

func (this *StreamsCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		this.persistence = spec.Spec.Configuration.Persistence
		this.bootstrapServers = spec.Spec.Configuration.Streams.BootstrapServers
		this.applicationServerPort = spec.Spec.Configuration.Streams.ApplicationServerPort
		this.applicationId = spec.Spec.Configuration.Streams.ApplicationId
		// TODO Use secrets!
		// Default values
		if this.applicationServerPort == "" {
			this.applicationServerPort = "9000"
		}
		if this.applicationId == "" {
			this.applicationId = spec.Name
		}
	}

	// Observation #2 + #3
	// Is the correct persistence type selected?
	// Validate the config values
	this.valid = this.persistence == "streams" && this.bootstrapServers != ""

	// Observation #4
	// Read the env values
	if val, exists := this.svcEnvCache.Get(ENV_KAFKA_BOOTSTRAP_SERVERS); exists {
		this.envBootstrapServers = val.GetValue().Value
	}
	_, exists := this.svcEnvCache.Get(ENV_APPLICATION_SERVER_HOST)
	this.envApplicationServerHostExists = exists
	if val, exists := this.svcEnvCache.Get(ENV_APPLICATION_SERVER_PORT); exists {
		this.envApplicationServerPort = val.GetValue().Value
	}
	if val, exists := this.svcEnvCache.Get(ENV_APPLICATION_ID); exists {
		this.envApplicationId = val.GetValue().Value
	}

	// We won't actively delete old env values if not used
}

func (this *StreamsCF) Compare() bool {
	// Condition #1
	// Is JPA & config values are valid
	// Condition #2 + #3
	// The required env vars are not present OR they differ
	return this.valid && (this.bootstrapServers != this.envBootstrapServers ||
		!this.envApplicationServerHostExists ||
		this.applicationServerPort != this.envApplicationServerPort ||
		this.applicationId != this.envApplicationId)
}

func (this *StreamsCF) Respond() {
	// Response #1
	// Just set the value(s)!
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_KAFKA_BOOTSTRAP_SERVERS, this.bootstrapServers))
	if !this.envApplicationServerHostExists {
		this.svcEnvCache.Set(env.NewEnvCacheEntry(&core.EnvVar{
			Name: ENV_APPLICATION_SERVER_HOST,
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		}))
	}
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_APPLICATION_SERVER_PORT, this.applicationServerPort))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_APPLICATION_ID, this.applicationId))

}

func (this *StreamsCF) Cleanup() bool {
	// No cleanup
	return true
}
