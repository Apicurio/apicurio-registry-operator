package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
)

var _ loop.ControlFunction = &KeycloakCF{}

const (
	ENV_REGISTRY_AUTH_ENABLED           = "AUTH_ENABLED"
	ENV_REGISTRY_KEYCLOAK_URL           = "KEYCLOAK_URL"
	ENV_REGISTRY_KEYCLOAK_REALM         = "KEYCLOAK_REALM"
	ENV_REGISTRY_KEYCLOAK_API_CLIENT_ID = "KEYCLOAK_API_CLIENT_ID"
	ENV_REGISTRY_KEYCLOAK_UI_CLIENT_ID  = "KEYCLOAK_UI_CLIENT_ID"

	DEFAULT_REGISTRY_KEYCLOAK_API_CLIENT_ID = "registry-client-api"
	DEFAULT_REGISTRY_KEYCLOAK_UI_CLIENT_ID  = "registry-client-ui"
)

type KeycloakCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	specEntry        resources.ResourceCacheEntry
	svcEnvCache      env.EnvCache
	valid            bool

	keycloakUrl         string
	keycloakRealm       string
	keycloakApiClientId string
	keycloakUiClientId  string

	envAuthEnabled         string
	envKeycloakUrl         string
	envKeycloakRealm       string
	envKeycloakApiClientId string
	envKeycloakUiClientId  string

	applyDefaultApi bool
	applyDefaultUi  bool
}

func NewKeycloakCF(ctx *context.LoopContext) loop.ControlFunction {
	return &KeycloakCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
	}
}

func (this *KeycloakCF) Describe() string {
	return "KeycloakCF"
}

func (this *KeycloakCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.specEntry = specEntry
		spec := specEntry.GetValue().(*ar.ApicurioRegistry).Spec
		this.keycloakUrl = spec.Configuration.Security.Keycloak.Url
		this.keycloakRealm = spec.Configuration.Security.Keycloak.Realm
		this.keycloakApiClientId = spec.Configuration.Security.Keycloak.ApiClientId
		this.keycloakUiClientId = spec.Configuration.Security.Keycloak.UiClientId
	} else {
		return
	}

	// Observation #2
	// Validate the config values
	this.valid = this.keycloakUrl != "" && this.keycloakRealm != ""

	// Observation #3
	// Read the env values
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_AUTH_ENABLED); exists {
		this.envAuthEnabled = val.GetValue().Value
	}
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_KEYCLOAK_URL); exists {
		this.envKeycloakUrl = val.GetValue().Value
	}
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_KEYCLOAK_REALM); exists {
		this.envKeycloakRealm = val.GetValue().Value
	}
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_KEYCLOAK_API_CLIENT_ID); exists {
		this.envKeycloakApiClientId = val.GetValue().Value
	}
	if val, exists := this.svcEnvCache.Get(ENV_REGISTRY_KEYCLOAK_UI_CLIENT_ID); exists {
		this.envKeycloakUiClientId = val.GetValue().Value
	}

	// Observation #4
	// Default values
	if this.keycloakApiClientId == "" {
		this.applyDefaultApi = true
		this.keycloakApiClientId = DEFAULT_REGISTRY_KEYCLOAK_API_CLIENT_ID
	}
	if this.keycloakUiClientId == "" {
		this.applyDefaultUi = true
		this.keycloakUiClientId = DEFAULT_REGISTRY_KEYCLOAK_UI_CLIENT_ID
	}
}

func (this *KeycloakCF) Compare() bool {
	// Condition #1
	return this.valid && (this.envAuthEnabled != "true" ||
		this.keycloakUrl != this.envKeycloakUrl ||
		this.keycloakRealm != this.envKeycloakRealm ||
		this.keycloakApiClientId != this.envKeycloakApiClientId ||
		this.keycloakUiClientId != this.envKeycloakUiClientId)
}

func (this *KeycloakCF) Respond() {
	// Response #1
	// Just set the value(s)!
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_AUTH_ENABLED, "true"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KEYCLOAK_URL, this.keycloakUrl))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KEYCLOAK_REALM, this.keycloakRealm))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KEYCLOAK_API_CLIENT_ID, this.keycloakApiClientId))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KEYCLOAK_UI_CLIENT_ID, this.keycloakUiClientId))

	// Response #1
	// Update defaults
	if this.applyDefaultApi {
		this.specEntry.ApplyPatch(func(value interface{}) interface{} {
			val := value.(*ar.ApicurioRegistry).DeepCopy()
			val.Spec.Configuration.Security.Keycloak.ApiClientId = this.keycloakApiClientId
			return val
		})
	}
	if this.applyDefaultUi {
		this.specEntry.ApplyPatch(func(value interface{}) interface{} {
			val := value.(*ar.ApicurioRegistry).DeepCopy()
			val.Spec.Configuration.Security.Keycloak.UiClientId = this.keycloakUiClientId
			return val
		})
	}
}

func (this *KeycloakCF) Cleanup() bool {
	// No cleanup
	return true
}
