package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

var _ loop.ControlFunction = &EnvCF{}

type EnvCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	// To know which were deleted, we need to compare with previous ones
	previousTargetEnv []corev1.EnvVar
	targetEnv         []corev1.EnvVar
	remove            map[string]corev1.EnvVar
	update            bool
}

// NewEnvCF creates a new instance of `Env` control function.
// This control function is responsible for reading custom environment variables from the spec,
// and saving them into the environment cache.
func NewEnvCF(ctx context.LoopContext) loop.ControlFunction {
	res := &EnvCF{
		ctx:               ctx,
		svcResourceCache:  ctx.GetResourceCache(),
		svcEnvCache:       ctx.GetEnvCache(),
		previousTargetEnv: make([]corev1.EnvVar, 0),
		targetEnv:         make([]corev1.EnvVar, 0),
		remove:            make(map[string]corev1.EnvVar),
		update:            false,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *EnvCF) Describe() string {
	return "EnvCF"
}

func (this *EnvCF) Sense() {
	this.log.Debugw("env cache before", "value", this.svcEnvCache.GetSorted())
	this.update = false
	this.remove = make(map[string]corev1.EnvVar)

	this.targetEnv = make([]corev1.EnvVar, 0)

	// Spec resource must be available
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		envConfig := specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Configuration.Env

		for _, v := range envConfig {
			this.targetEnv = append(this.targetEnv, v)
			if _, e := this.svcEnvCache.Get(v.Name); !e {
				// Add to cache
				this.update = true
			}
		}

		this.remove = make(map[string]corev1.EnvVar, 0)
		for _, v := range this.svcEnvCache.GetSorted() { // TODO Iterate directly?
			cached, _ := this.svcEnvCache.Get(v.Name)
			// Iterate over cached
			if !found(envConfig, v.Name) && cached.GetPriority() == env.PRIORITY_SPEC {
				// Element is in the cache, but not in spec
				this.remove[v.Name] = v
			}
		}

		// Update even when the env. variables have been reordered.
		// This is important in case of variable interpolation
		if len(this.previousTargetEnv) == len(this.targetEnv) {
			for i, _ := range this.targetEnv {
				if !reflect.DeepEqual(this.targetEnv[i], this.previousTargetEnv[i]) {
					this.update = true
					break
				}
			}
		} else {
			this.update = true
		}
	}
}

func (this *EnvCF) Compare() bool {
	this.log.Debugw("conditions", "this.update", this.update, "len(this.remove)", len(this.remove))
	return this.update || len(this.remove) > 0
}

func (this *EnvCF) Respond() {
	// Response #1
	// Remove first
	for _, v := range this.remove {
		this.svcEnvCache.DeleteByName(v.Name)
	}

	// Response #2
	// We do not update changed variables only
	// to keep the ordering of the values as defined in spec.
	prev := ""
	for _, v := range this.targetEnv {
		// Add to the cache (overwrite)
		entryBuilder := env.NewEnvCacheEntryBuilder(&v)
		if prev != "" {
			// Maintain ordering
			entryBuilder.SetDependency(prev)
		}
		this.svcEnvCache.Set(entryBuilder.SetPriority(env.PRIORITY_SPEC).Build())
		prev = v.Name
	}

	this.previousTargetEnv = this.targetEnv
	this.log.Debugw("env cache after", "value", this.svcEnvCache.GetSorted())
}

func (this *EnvCF) Cleanup() bool {
	// No cleanup
	return true
}

func found(haystack []corev1.EnvVar, needle string) bool {
	for _, v := range haystack {
		if v.Name == needle {
			return true
		}
	}
	return false
}
