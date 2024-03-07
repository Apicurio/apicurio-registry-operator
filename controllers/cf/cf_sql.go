package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
)

var _ loop.ControlFunction = &SqlCF{}

const ENV_REGISTRY_DATASOURCE_URL = "REGISTRY_DATASOURCE_URL"
const ENV_REGISTRY_DATASOURCE_USERNAME = "REGISTRY_DATASOURCE_USERNAME"
const ENV_REGISTRY_DATASOURCE_PASSWORD = "REGISTRY_DATASOURCE_PASSWORD"

type SqlCF struct {
	ctx              context.LoopContext
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	persistence      string
	valid            bool
	url              string
	envUrl           env.EnvCacheEntry
	user             string
	envUser          env.EnvCacheEntry
	password         string
	envPassword      env.EnvCacheEntry
	log              *zap.SugaredLogger
}

func NewSqlCF(ctx context.LoopContext) loop.ControlFunction {
	return &SqlCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		persistence:      "",
		valid:            true,
		url:              "",
		envUrl:           nil,
		user:             "",
		envUser:          nil,
		password:         "",
		envPassword:      nil,
		log:              ctx.GetLog().Sugar(),
	}
}

func (this *SqlCF) Describe() string {
	return "SqlCF"
}

func (this *SqlCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry).Spec
		this.persistence = spec.Configuration.Persistence
		this.url = spec.Configuration.Sql.DataSource.Url
		this.user = spec.Configuration.Sql.DataSource.UserName
		this.password = spec.Configuration.Sql.DataSource.Password // Leave empty as default
		// TODO Use secrets!
	}

	// Observation #2
	// Is the correct persistence type selected?
	// Validate the config values
	this.valid = this.persistence == "sql" && (this.url != "" || this.envUrl != nil) && (this.user != "" || this.envUser != nil)

	// Observation #3
	// Read the env values
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_DATASOURCE_URL); exists {
		this.envUrl = val
	} else {
		this.envUrl = nil
	}
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_DATASOURCE_USERNAME); exists {
		this.envUser = val
	} else {
		this.envUser = nil
	}
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_DATASOURCE_PASSWORD); exists {
		this.envPassword = val
	} else {
		this.envPassword = nil
	}
}

func (this *SqlCF) Compare() bool {
	var updateUrl = false
	if this.envUrl == nil {
		updateUrl = this.url != ""
	} else {
		// Values differ, and either we override the data from spec.configuration.env (if spec...url is set), or we have created the variable ourselves and need to update it accordingly
		updateUrl = this.url != this.envUrl.GetValue().Value && (this.url != "" || this.envUrl.GetPriority() == env.PRIORITY_OPERATOR)
	}
	var updateUser = false
	if this.envUser == nil {
		updateUser = this.user != ""
	} else {
		updateUser = this.user != this.envUser.GetValue().Value && (this.user != "" || this.envUser.GetPriority() == env.PRIORITY_OPERATOR)
	}
	var updatePassword = false
	if this.envPassword == nil {
		updatePassword = true // Password can be empty
	} else {
		updatePassword = this.password != this.envPassword.GetValue().Value && (this.password != "" || this.envPassword.GetPriority() == env.PRIORITY_OPERATOR || this.envPassword.GetPriority() == env.PRIORITY_MIN)
	}
	return this.valid && (updateUrl || updateUser || updatePassword)
}

func (this *SqlCF) Respond() {

	if this.url != "" {
		// Not empty, we just set the variable, overriding spec.configuration.env
		this.svcEnvCache.Set(env.NewSimpleEnvCacheEntryBuilder(ENV_REGISTRY_DATASOURCE_URL, this.url).Build())
	} else {
		if this.envUrl != nil {
			if this.envUrl.GetPriority() == env.PRIORITY_OPERATOR {
				// We've set it, we can delete it
				this.svcEnvCache.DeleteByName(ENV_REGISTRY_DATASOURCE_URL)
			} // else is managed by spec.configuration.env
		} else {
			// Invalid state
			panic("unreachable")
		}
	}

	if this.user != "" {
		this.svcEnvCache.Set(env.NewSimpleEnvCacheEntryBuilder(ENV_REGISTRY_DATASOURCE_USERNAME, this.user).Build())
	} else {
		if this.envUser != nil {
			if this.envUser.GetPriority() == env.PRIORITY_OPERATOR {
				this.svcEnvCache.DeleteByName(ENV_REGISTRY_DATASOURCE_USERNAME)
			}
		} else {
			panic("unreachable")
		}
	}

	if this.password != "" {
		// Not empty, we just set the variable, overriding spec.configuration.env
		this.svcEnvCache.Set(env.NewSimpleEnvCacheEntryBuilder(ENV_REGISTRY_DATASOURCE_PASSWORD, this.password).Build())
	} else {
		// Set empty password, but make it overridable
		this.svcEnvCache.Set(env.NewSimpleEnvCacheEntryBuilder(ENV_REGISTRY_DATASOURCE_PASSWORD, this.password).SetPriority(env.PRIORITY_MIN).Build())
	}
}

func (this *SqlCF) Cleanup() bool {
	// No cleanup
	return true
}
