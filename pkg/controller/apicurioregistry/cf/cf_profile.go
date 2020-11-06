package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
)

var _ loop.ControlFunction = &ProfileCF{}

const ENV_QUARKUS_PROFILE = "QUARKUS_PROFILE"

type ProfileCF struct {
	ctx         *context.LoopContext
	svcEnvCache env.EnvCache
	profileSet  bool
}

// Is responsible for managing environment variables from the env cache
func NewProfileCF(ctx *context.LoopContext) loop.ControlFunction {
	return &ProfileCF{
		ctx:         ctx,
		svcEnvCache: ctx.GetEnvCache(),
		profileSet:  false,
	}
}

func (this *ProfileCF) Describe() string {
	return "ProfileCF"
}

func (this *ProfileCF) Sense() {
	// Observation #1
	// Was the profile env var set?
	_, profileSet := this.svcEnvCache.Get(ENV_QUARKUS_PROFILE)
	this.profileSet = profileSet

}

func (this *ProfileCF) Compare() bool {
	// Condition #1
	// Env var does not exist
	return !this.profileSet
}

func (this *ProfileCF) Respond() {
	// Response #1
	// Just set the value(s)!
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_QUARKUS_PROFILE, "prod"))

}

func (this *ProfileCF) Cleanup() bool {
	// No cleanup
	return true
}
