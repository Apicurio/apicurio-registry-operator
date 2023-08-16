package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	"net/url"
)

var _ loop.ControlFunction = &CorsCF{}

const ENV_CORS = "CORS_ALLOWED_ORIGINS"

type CorsCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache
	targetCors       string
	existingCors     string
	overriddenCors   string
}

// This CF makes sure the CORS_ALLOWED_ORIGINS env. variable is set properly
func NewCorsCF(ctx context.LoopContext) loop.ControlFunction {
	res := &CorsCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *CorsCF) Describe() string {
	return "CorsCF"
}

func (this *CorsCF) Sense() {

	this.targetCors = ""
	this.overriddenCors = ""
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		host := specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
		if host != "" {
			this.targetCors = "http://" + host + "," + "https://" + host
		}
		keycloak := specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Configuration.Security.Keycloak.Url
		if keycloak != "" {
			if keycloakUrl, err := url.Parse(keycloak); err == nil {
				if keycloakHost := keycloakUrl.Hostname(); keycloakHost != "" {
					if this.targetCors != "" {
						this.targetCors = this.targetCors + ","
					}
					this.targetCors = this.targetCors + keycloakUrl.Scheme + "://" + keycloakHost
				} else {
					this.log.With("keycloakUrl", keycloakUrl).
						Infof("could not include Keycloak URL in %s, failed to get host. "+
							"Make sure the URL is a valid URL with both a scheme and a host", ENV_CORS)
				}
			} else {
				this.log.With("keycloakUrl", keycloakUrl, "error", err).
					Infof("could not include Keycloak URL in %s, failed to parse URL", ENV_CORS)
			}
		}

		envList := specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Configuration.Env
		for _, e := range envList {
			if e.Name == ENV_CORS {
				this.overriddenCors = e.Value
			}
		}
	}

	this.existingCors = ""
	if entry, exists := this.ctx.GetEnvCache().Get(ENV_CORS); exists {
		this.existingCors = entry.GetValue().Value
	}
}

func (this *CorsCF) Compare() bool {

	if this.overriddenCors != "" {
		return this.existingCors != this.overriddenCors
	} else {
		return this.existingCors != this.targetCors
	}
}

func (this *CorsCF) Respond() {

	if this.overriddenCors != "" {
		this.ctx.GetEnvCache().Set(env.NewSimpleEnvCacheEntryBuilder(ENV_CORS, this.overriddenCors).Build())
	} else {
		if this.targetCors != "" {
			this.ctx.GetEnvCache().Set(env.NewSimpleEnvCacheEntryBuilder(ENV_CORS, this.targetCors).Build())
		} else {
			this.ctx.GetEnvCache().DeleteByName(ENV_CORS)
		}
	}
}

func (this *CorsCF) Cleanup() bool {
	// No cleanup
	return true
}
