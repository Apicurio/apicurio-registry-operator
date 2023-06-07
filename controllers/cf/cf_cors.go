package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
)

var _ loop.ControlFunction = &CorsCF{}

const ENV_CORS = "CORS_ALLOWED_ORIGINS"

type CorsCF struct {
	ctx              context.LoopContext
	svcResourceCache resources.ResourceCache
	targetCors       string
	existingCors     string
	overridden       bool
}

// This CF makes sure the CORS_ALLOWED_ORIGINS env. variable is set properly
func NewCorsCF(ctx context.LoopContext) loop.ControlFunction {
	return &CorsCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		targetCors:       "",
	}
}

func (this *CorsCF) Describe() string {
	return "CorsCF"
}

func (this *CorsCF) Sense() {

	this.targetCors = ""
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		host := specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
		if host != "" {
			this.targetCors = "http://" + host + "," +
				"https://" + host
		}
	}

	this.existingCors = ""
	if entry, exists := this.ctx.GetEnvCache().Get(ENV_CORS); exists {
		this.existingCors = entry.GetValue().Value
		this.overridden = entry.GetPriority() != env.PRIORITY_OPERATOR
	}
}

func (this *CorsCF) Compare() bool {

	return this.existingCors != this.targetCors && !this.overridden
}

func (this *CorsCF) Respond() {

	if this.targetCors != "" {
		this.ctx.GetEnvCache().Set(env.NewSimpleEnvCacheEntryBuilder(ENV_CORS, this.targetCors).Build())
	} else {
		this.ctx.GetEnvCache().DeleteByName(ENV_CORS)
	}
}

func (this *CorsCF) Cleanup() bool {
	// No cleanup
	return true
}
