package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
)

var _ loop.ControlFunction = &JpaCF{}

const ENV_QUARKUS_DATASOURCE_URL = "QUARKUS_DATASOURCE_URL"
const ENV_QUARKUS_DATASOURCE_USERNAME = "QUARKUS_DATASOURCE_USERNAME"
const ENV_QUARKUS_DATASOURCE_PASSWORD = "QUARKUS_DATASOURCE_PASSWORD"

type JpaCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	persistence      string
	url              string
	user             string
	password         string
	valid            bool
	envUrl           string
	envUser          string
	envPassword      string
}

func NewJpaCF(ctx *context.LoopContext) loop.ControlFunction {
	return &JpaCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		persistence:      "",
		url:              "",
		user:             "",
		password:         "",
		valid:            true,
		envUrl:           "",
		envUser:          "",
		envPassword:      "",
	}
}

func (this *JpaCF) Describe() string {
	return "JpaCF"
}

func (this *JpaCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry).Spec
		this.persistence = spec.Configuration.Persistence
		this.url = spec.Configuration.DataSource.Url
		this.user = spec.Configuration.DataSource.UserName
		this.password = spec.Configuration.DataSource.Password // Leave empty as default
		// TODO Use secrets!
	}

	// Observation #2 + #3
	// Is the correct persistence type selected?
	// Validate the config values
	this.valid = this.persistence == "jpa" && this.url != "" && this.user != ""

	// Observation #4
	// Read the env values
	if val, exists := this.svcEnvCache.Get(ENV_QUARKUS_DATASOURCE_URL); exists {
		this.envUrl = val.GetValue().Value
	}
	if val, exists := this.svcEnvCache.Get(ENV_QUARKUS_DATASOURCE_USERNAME); exists {
		this.envUser = val.GetValue().Value
	}
	if val, exists := this.svcEnvCache.Get(ENV_QUARKUS_DATASOURCE_PASSWORD); exists {
		this.envPassword = val.GetValue().Value
	}

	// We won't actively delete old env values if not used
}

func (this *JpaCF) Compare() bool {
	// Condition #1
	// Is JPA & config values are valid
	// Condition #2 + #3
	// The required env vars are not present OR they differ
	return this.valid && (this.url != this.envUrl ||
		this.user != this.envUser ||
		this.password != this.envPassword)
}

func (this *JpaCF) Respond() {
	// Response #1
	// Just set the value(s)!
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_QUARKUS_DATASOURCE_URL, this.url))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_QUARKUS_DATASOURCE_USERNAME, this.user))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_QUARKUS_DATASOURCE_PASSWORD, this.password))

}

func (this *JpaCF) Cleanup() bool {
	// No cleanup
	return true
}
