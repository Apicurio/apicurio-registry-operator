package apicurioregistry

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
)

type Patcher struct {
	ctx *Context
	// ===
	deploymentUFs []DeploymentUF
	serviceUFs    []ServiceUF
	ingressUFs    []IngressUF
}

// Patcher provides an easy way to modify resources that are tracked by this operator
func NewPatcher(ctx *Context) *Patcher {
	return &Patcher{
		ctx: ctx,
	}
}

// ===

// Add a function that will be executed to modify a Deployment
func (this *Patcher) AddDeploymentPatch(updateFunc DeploymentUF) {
	this.deploymentUFs = append(this.deploymentUFs, updateFunc)
}

func (this *Patcher) patchDeployment() {
	if name := this.ctx.configuration.GetConfig(CFG_STA_DEPLOYMENT_NAME); name != "" {
		err := this.ctx.kubecl.PatchDeployment(this.ctx.configuration.GetSpecNamespace(), name, func(deployment *apps.Deployment) {
			for _, v := range this.deploymentUFs {
				v(deployment)
			}
		})
		if err != nil {
			this.ctx.log.Error(err, "Error during Deployment patching")
		}
	}
}

// ===

// Add a function that will be executed to modify a Service
func (this *Patcher) AddServicePatch(updateFunc ServiceUF) {
	this.serviceUFs = append(this.serviceUFs, updateFunc)
}

func (this *Patcher) patchService() {
	if name := this.ctx.configuration.GetConfig(CFG_STA_SERVICE_NAME); name != "" {
		err := this.ctx.kubecl.PatchService(this.ctx.configuration.GetSpecNamespace(), name, func(service *core.Service) {
			for _, v := range this.serviceUFs {
				v(service)
			}
		})
		if err != nil {
			this.ctx.log.Error(err, "Error during Service patching")
		}
	}
}

// ===

// Add a function that will be executed to modify an Ingress
func (this *Patcher) AddIngressPatch(updateFunc IngressUF) {
	this.ingressUFs = append(this.ingressUFs, updateFunc)
}

func (this *Patcher) patchIngress() {
	if name := this.ctx.configuration.GetConfig(CFG_STA_INGRESS_NAME); name != "" {
		err := this.ctx.kubecl.PatchIngress(this.ctx.configuration.GetSpecNamespace(), name, func(ingress *extensions.Ingress) {
			for _, v := range this.ingressUFs {
				v(ingress)
			}
		})
		if err != nil {
			this.ctx.log.Error(err, "Error during Ingress patching")
		}
	}
}

// ===

// Do the patching
func (this *Patcher) Execute() {
	this.patchDeployment()
	this.patchService()
	this.patchIngress()
	// reset
	this.deploymentUFs = *new([]DeploymentUF)
	this.serviceUFs = *new([]ServiceUF)
	this.ingressUFs = *new([]IngressUF)
}
