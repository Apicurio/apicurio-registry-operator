package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
)

var _ loop.ControlFunction = &EnvApplyCF{}

type EnvApplyCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache

	svcEnvCache     env.EnvCache
	envCacheUpdated bool

	deploymentExists      bool
	deploymentEntry       resources.ResourceCacheEntry
	deploymentNeedsUpdate bool
}

// Is responsible for managing environment variables from the env cache
func NewEnvApplyCF(ctx context.LoopContext) loop.ControlFunction {
	res := &EnvApplyCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),

		svcEnvCache:     ctx.GetEnvCache(),
		envCacheUpdated: false,

		deploymentExists: false,
		deploymentEntry:  nil,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *EnvApplyCF) Describe() string {
	return "EnvApplyCF"
}

func (this *EnvApplyCF) Sense() {
	this.log.Debugw("env cache before", "value", this.svcEnvCache.GetSorted())
	// Observation #1
	// Is deployment available and/or is it already created
	var deploymentEntry resources.ResourceCacheEntry
	if deploymentEntry, this.deploymentExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); this.deploymentExists {
		this.deploymentEntry = deploymentEntry
		deployment := this.deploymentEntry.GetValue().(*apps.Deployment)

		// Observation #2
		// Determine whether any env. variables are missing or need to be removed
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == factory.REGISTRY_CONTAINER_NAME {

				this.deploymentNeedsUpdate = false

				// Remember which variables are in the cache, to see if any is NOT present in the deployment
				cachedVariablePresentInDeployment := make(map[string]bool, 0)
				for _, v := range this.svcEnvCache.GetSorted() {
					if v.Name == env.JAVA_OPTIONS_OPERATOR || v.Name == env.JAVA_OPTIONS_COMBINED {
						continue // Ignore, these are special internal-only env. variables.
					}
					cachedVariablePresentInDeployment[v.Name] = false
				}
				for _, e := range deployment.Spec.Template.Spec.Containers[i].Env {
					// Mark as present:
					cachedVariablePresentInDeployment[e.Name] = true
					// Also ensure the other direction, each variable in deployment is present in the cache:
					if _, exists := this.svcEnvCache.Get(e.Name); !exists {
						_, combinedExists := this.svcEnvCache.Get(env.JAVA_OPTIONS_COMBINED)
						if e.Name == env.JAVA_OPTIONS && combinedExists {
							// We transform env.JAVA_OPTIONS_COMBINED into env.JAVA_OPTIONS, which may not be present in the cache beforehand.
							this.log.Debugln("ignoring that variable " + env.JAVA_OPTIONS + " is in deployment but not in cache")
						} else {
							this.log.Debugln("variable is in deployment but not in cache", e.Name)
							this.deploymentNeedsUpdate = true
						}
					}
				}
				// Check that all variables in the cache are present in the deployment:
				for k, v := range cachedVariablePresentInDeployment {
					if !v {
						this.log.Debugln("variable is in cache but not in deployment", k)
						this.deploymentNeedsUpdate = true
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
	this.log.Debugw("env apply compare", "this.envCacheUpdated", this.envCacheUpdated,
		"this.deploymentExists", this.deploymentExists,
		"this.deploymentNeedsUpdate", this.deploymentNeedsUpdate,
	)
	return this.deploymentExists && (this.envCacheUpdated || this.deploymentNeedsUpdate)
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

				this.log.Debugw("deployment env after", "env", sorted)
				deployment.Spec.Template.Spec.Containers[i].Env = sorted
			}
		} // TODO report a problem if not found?
		return deployment
	})

	// Response #2
	// Do not clear the cache, but reset the change mark
	this.svcEnvCache.ProcessAndAdvanceToNextPeriod()

}

func (this *EnvApplyCF) Cleanup() bool {
	// No cleanup
	return true
}
