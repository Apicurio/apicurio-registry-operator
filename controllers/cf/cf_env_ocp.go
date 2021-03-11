package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	ocp_apps "github.com/openshift/api/apps/v1"
)

var _ loop.ControlFunction = &EnvOcpCF{}

type EnvOcpCF struct {
	ctx                *context.LoopContext
	svcResourceCache   resources.ResourceCache
	svcEnvCache        env.EnvCache
	deploymentExists   bool
	deploymentEntry    resources.ResourceCacheEntry
	deploymentName     string
	lastDeploymentName string
	envCacheUpdated    bool
}

// Is responsible for managing environment variables from the env cache
func NewEnvOcpCF(ctx *context.LoopContext) loop.ControlFunction {
	return &EnvOcpCF{
		ctx:                ctx,
		svcResourceCache:   ctx.GetResourceCache(),
		svcEnvCache:        ctx.GetEnvCache(),
		deploymentExists:   false,
		deploymentEntry:    nil,
		deploymentName:     resources.RC_EMPTY_NAME,
		lastDeploymentName: resources.RC_EMPTY_NAME,
		envCacheUpdated:    false,
	}
}

func (this *EnvOcpCF) Describe() string {
	return "EnvOcpCF"
}

func (this *EnvOcpCF) Sense() {
	// Observation #1
	// Is deployment available and/or is it already created
	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT_OCP)
	this.deploymentExists = deploymentExists
	this.deploymentEntry = deploymentEntry
	this.deploymentName = deploymentEntry.GetName().Str()

	// Observation #2
	// Was the env cache updated?
	this.envCacheUpdated = this.svcEnvCache.IsChanged()

}

func (this *EnvOcpCF) Compare() bool {
	// Condition #1
	// We have something to update
	// Condition #2
	// There is a deployment
	return (this.envCacheUpdated || this.deploymentName != this.lastDeploymentName) && this.deploymentExists
}

func (this *EnvOcpCF) Respond() {
	// Response #1
	// First, read the existing env variables, and the add them to cache,
	// so they stay at the end where possible, keeping the order where possible, because
	// we do not have dependency info about those.
	// The operator overwrites user defined ones only when necessary
	deployment := this.deploymentEntry.GetValue().(*ocp_apps.DeploymentConfig)
	for i, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == this.ctx.GetAppName().Str() {
			for _, e := range deployment.Spec.Template.Spec.Containers[i].Env {
				// Add to the cache
				if v, exists := this.svcEnvCache.Get(e.Name); exists {
					if !v.IsManaged() { // TODO this avoids overwriting of managed env variables
						this.svcEnvCache.Set(env.NewEnvCacheEntryUnmanaged(e.DeepCopy()))
					}
				} else {
					this.svcEnvCache.Set(env.NewEnvCacheEntryUnmanaged(e.DeepCopy()))
				}
			}
		}
	} // TODO report a problem if not found?

	// Response #2
	// Write the sorted env vars
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*ocp_apps.DeploymentConfig).DeepCopy()
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				deployment.Spec.Template.Spec.Containers[i].Env = this.svcEnvCache.GetSorted()
			}
		} // TODO report a problem if not found?
		return deployment
	})

	// Response #3
	// Do not clear the cache, but reset the change mark
	this.svcEnvCache.ResetChanged()

	this.lastDeploymentName = this.deploymentName
}

func (this *EnvOcpCF) Cleanup() bool {
	// No cleanup
	return true
}
