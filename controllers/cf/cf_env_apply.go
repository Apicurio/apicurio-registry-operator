package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ loop.ControlFunction = &EnvApplyCF{}

type EnvApplyCF struct {
	ctx                context.LoopContext
	log                *zap.SugaredLogger
	svcResourceCache   resources.ResourceCache
	svcEnvCache        env.EnvCache
	deploymentExists   bool
	deploymentEntry    resources.ResourceCacheEntry
	deploymentName     string
	envCacheUpdated    bool
	lastDeploymentName string
	deploymentUID      types.UID
	lastDeploymentUID  types.UID
}

// Is responsible for managing environment variables from the env cache
func NewEnvApplyCF(ctx context.LoopContext) loop.ControlFunction {
	res := &EnvApplyCF{
		ctx:                ctx,
		svcResourceCache:   ctx.GetResourceCache(),
		svcEnvCache:        ctx.GetEnvCache(),
		deploymentExists:   false,
		deploymentEntry:    nil,
		deploymentName:     resources.RC_NOT_CREATED_NAME_EMPTY,
		lastDeploymentName: resources.RC_NOT_CREATED_NAME_EMPTY,
		envCacheUpdated:    false,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *EnvApplyCF) Describe() string {
	return "EnvApplyCF"
}

func (this *EnvApplyCF) Sense() {
	// Observation #1
	// Is deployment available and/or is it already created
	var deploymentEntry resources.ResourceCacheEntry
	if deploymentEntry, this.deploymentExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); this.deploymentExists {
		this.deploymentEntry = deploymentEntry
		this.deploymentName = deploymentEntry.GetName().Str()

		// Observation #2
		// First, read the existing env variables, and the add them to cache,
		// keeping the original ordering.
		// The operator overwrites user defined ones only when necessary.
		deployment := this.deploymentEntry.GetValue().(*apps.Deployment)

		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == factory.REGISTRY_CONTAINER_NAME {
				prevName := "" // To maintain ordering in case of interpolation
				// Copy variables in the cache
				deleted := make(map[string]bool, 0)
				for _, v := range this.svcEnvCache.GetSorted() {
					if v.Name == env.JAVA_OPTIONS_OPERATOR || v.Name == env.JAVA_OPTIONS_COMBINED {
						continue // Do not delete, these are special internal-only env. variables.
						// TODO: Consider refactoring this so the special case is no longer needed.
					}
					deleted[v.Name] = true // deletes spec stuff as well
				}
				for _, e := range deployment.Spec.Template.Spec.Containers[i].Env {

					// Remove from deleted if in spec
					delete(deleted, e.Name)

					// If already marked as deleted, do not re-add them
					if this.svcEnvCache.WasDeleted(e.Name) {
						continue
					}

					if e.Name == env.JAVA_OPTIONS {
						/*
							Do not read this variable from deployment, we can't support this specific use-case anymore.
							It has been deprecated for some time, and it's very unlikely someone relies on this behavior.
						*/
						continue
					}

					// Add to the cache
					entryBuilder := env.NewEnvCacheEntryBuilder(&e).SetPriority(env.PRIORITY_DEPLOYMENT)
					if prevName != "" {
						entryBuilder.SetDependency(prevName)
					}
					this.svcEnvCache.Set(entryBuilder.Build())
					prevName = e.Name
				}
				// Remove things from the cache that are not in the spec
				// IF the cache was not changed already.
				// This would otherwise prevent new things from being added.
				if !this.svcEnvCache.IsChanged() {
					for k, _ := range deleted {
						this.svcEnvCache.DeleteByName(k)
					}
				}
			}
		}
	}

	// Handle Java Options legacy variable by parsing and saving again ¯\_(ツ)_/¯
	if parsed, err := env.ParseCombinedJavaOptionsMap(this.svcEnvCache); err == nil {
		env.SaveCombinedJavaOptionsMap(this.svcEnvCache, parsed)
	} else {
		this.log.Errorw("could not parse env. variables "+env.JAVA_OPTIONS+" or "+env.JAVA_OPTIONS_LEGACY, "error", err)
	}

	// Observation #2
	// Was the env cache updated?
	this.envCacheUpdated = this.svcEnvCache.IsChanged()
}

func (this *EnvApplyCF) Compare() bool {
	// Condition #1
	// We have something to update
	// Condition #2
	// There is a deployment
	return (this.envCacheUpdated || this.deploymentName != this.lastDeploymentName) && this.deploymentExists
}

func (this *EnvApplyCF) Respond() {
	// Response #1
	// Write the sorted env vars
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == factory.REGISTRY_CONTAINER_NAME {

				sorted := this.svcEnvCache.GetSorted()
				// Hack to replace JAVA_OPTIONS with JAVA_OPTIONS_COMBINED
				if _, found := env.GetEnv(sorted, env.JAVA_OPTIONS_COMBINED); found {
					sorted, _ = env.RemoveEnv(sorted, env.JAVA_OPTIONS)
					sorted, _ = env.RemoveEnv(sorted, env.JAVA_OPTIONS_OPERATOR)
					v, _ := env.GetEnv(sorted, env.JAVA_OPTIONS_COMBINED)
					v.Name = env.JAVA_OPTIONS
				}
				deployment.Spec.Template.Spec.Containers[i].Env = sorted
			}
		} // TODO report a problem if not found?
		return deployment
	})

	// Response #2
	// Do not clear the cache, but reset the change mark
	this.svcEnvCache.ProcessAndAdvanceToNextPeriod()

	this.lastDeploymentName = this.deploymentName
}

func (this *EnvApplyCF) Cleanup() bool {
	// No cleanup
	return true
}
