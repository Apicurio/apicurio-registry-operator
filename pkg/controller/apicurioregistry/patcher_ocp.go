package apicurioregistry

import (
	ocp_apps "github.com/openshift/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type DeploymentOCPUF = func(spec *ocp_apps.DeploymentConfig)

type OCPPatcher struct {
	ctx           *Context
	deploymentUFs []DeploymentOCPUF
}

func NewOCPPatcher(ctx *Context) *OCPPatcher {
	return &OCPPatcher{
		ctx: ctx,
	}
}

// ===

func (this *OCPPatcher) AddDeploymentPatch(updateFunc DeploymentOCPUF) {
	this.deploymentUFs = append(this.deploymentUFs, updateFunc)
}

func (this *OCPPatcher) patchDeployment() {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_DEPLOYMENT_NAME); name != "" {
		err := this.ctx.GetClients().OCP().PatchDeployment(this.ctx.GetConfiguration().GetAppNamespace(), name, func(deployment *ocp_apps.DeploymentConfig) {
			for _, v := range this.deploymentUFs {
				v(deployment)
			}
		})
		if err != nil {
			this.ctx.GetLog().Error(err, "Error during Deployment patching")
		}
	}
}

// =====

func (this *OCPClient) PatchDeployment(namespace string, name string, updateFunc DeploymentOCPUF) error {
	o, err := this.GetDeployment(namespace, name, &meta.GetOptions{})
	if err != nil {
		return err
	}
	n := o.DeepCopy()
	updateFunc(n)
	patchData, err := createPatch(o, n, ocp_apps.DeploymentConfig{})
	if err != nil {
		return err
	}
	_, err = this.ocpAppsClient.DeploymentConfigs(namespace).Patch(name, types.StrategicMergePatchType, patchData)
	return err
}

func (this *OCPPatcher) Execute() {
	if len(this.deploymentUFs) > 0 {
		this.patchDeployment()
	}
	// reset
	this.deploymentUFs = *new([]DeploymentOCPUF)
}
