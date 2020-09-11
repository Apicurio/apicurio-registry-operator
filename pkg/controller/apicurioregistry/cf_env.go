package apicurioregistry

import (
	apps "k8s.io/api/apps/v1"
)

var _ ControlFunction = &EnvCF{}

type EnvCF struct {
	ctx                *Context
	deploymentExists   bool
	deploymentEntry    ResourceCacheEntry
	deploymentName     string
	envCacheUpdated    bool
	lastDeploymentName string
}

// Is responsible for managing environment variables from the env cache
func NewEnvCF(ctx *Context) ControlFunction {
	return &EnvCF{
		ctx:                ctx,
		deploymentExists:   false,
		deploymentEntry:    nil,
		deploymentName:     RC_EMPTY_NAME,
		lastDeploymentName: RC_EMPTY_NAME,
		envCacheUpdated:    false,
	}
}

func (this *EnvCF) Describe() string {
	return "EnvCF"
}

func (this *EnvCF) Sense() {
	// Observation #1
	// Is deployment available and/or is it already created
	deploymentEntry, deploymentExists := this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT)
	this.deploymentExists = deploymentExists
	this.deploymentEntry = deploymentEntry
	this.deploymentName = deploymentEntry.GetName()

	// Observation #2
	// Was the env cache updated?
	this.envCacheUpdated = this.ctx.GetEnvCache().IsChanged()

}

func (this *EnvCF) Compare() bool {
	// Condition #1
	// We have something to update
	// Condition #2
	// There is a deployment
	return (this.envCacheUpdated || this.deploymentName != this.lastDeploymentName) && this.deploymentExists
}

func (this *EnvCF) Respond() {
	// Response #1
	// First, read the existing env variables, and the add them to cache,
	// so they stay at the end where possible, keeping the order where possible, because
	// we do not have dependency info about those.
	// The operator overwrites user defined ones only when necessary
	deployment := this.deploymentEntry.GetValue().(*apps.Deployment)
	for i, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == this.ctx.GetConfiguration().GetAppName() {
			for _, e := range deployment.Spec.Template.Spec.Containers[i].Env {
				// Add to the cache
				if v, exists := this.ctx.GetEnvCache().Get(e.Name); exists {
					if !v.IsManaged() { // TODO this avoids overwriting of managed env variables
						this.ctx.GetEnvCache().Set(NewEnvCacheEntryUnmanaged(e.DeepCopy()))
					}
				} else {
					this.ctx.GetEnvCache().Set(NewEnvCacheEntryUnmanaged(e.DeepCopy()))
				}
			}
		}
	} // TODO report a problem if not found?

	// Response #2
	// Write the sorted env vars
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetConfiguration().GetAppName() {
				deployment.Spec.Template.Spec.Containers[i].Env = this.ctx.GetEnvCache().GetSorted()
			}
		} // TODO report a problem if not found?
		return deployment
	})

	// Response #3
	// Do not clear the cache, but reset the change mark
	this.ctx.GetEnvCache().ResetChanged()

	this.lastDeploymentName = this.deploymentName
}

func (this *EnvCF) Cleanup() bool {
	// No cleanup
	return true
}
