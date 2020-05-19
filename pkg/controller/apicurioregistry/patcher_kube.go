package apicurioregistry

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type DeploymentUF = func(spec *apps.Deployment)
type ServiceUF = func(spec *core.Service)
type IngressUF = func(spec *extensions.Ingress)

type KubePatcher struct {
	ctx           *Context
	deploymentUFs []DeploymentUF
	serviceUFs    []ServiceUF
	ingressUFs    []IngressUF
}

func NewKubePatcher(ctx *Context) *KubePatcher {
	return &KubePatcher{
		ctx: ctx,
	}
}

// ===

func (this *KubePatcher) AddDeploymentPatch(updateFunc DeploymentUF) {
	this.deploymentUFs = append(this.deploymentUFs, updateFunc)
}

func (this *KubePatcher) patchDeployment() {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_DEPLOYMENT_NAME); name != "" {
		err := this.ctx.GetClients().Kube().PatchDeployment(this.ctx.GetConfiguration().GetAppNamespace(), name, func(deployment *apps.Deployment) {
			for _, v := range this.deploymentUFs {
				v(deployment)
			}
		})
		if err != nil {
			this.ctx.GetLog().Error(err, "Error during Deployment patching")
		}
	}
}

// ===

func (this *KubePatcher) AddServicePatch(updateFunc ServiceUF) {
	this.serviceUFs = append(this.serviceUFs, updateFunc)
}

func (this *KubePatcher) patchService() {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME); name != "" {
		err := this.ctx.GetClients().Kube().PatchService(this.ctx.GetConfiguration().GetAppNamespace(), name, func(service *core.Service) {
			for _, v := range this.serviceUFs {
				v(service)
			}
		})
		if err != nil {
			this.ctx.GetLog().Error(err, "Error during Service patching")
		}
	}
}

// ===

func (this *KubePatcher) AddIngressPatch(updateFunc IngressUF) {
	this.ingressUFs = append(this.ingressUFs, updateFunc)
}

func (this *KubePatcher) patchIngress() {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_INGRESS_NAME); name != "" {
		err := this.ctx.GetClients().Kube().PatchIngress(this.ctx.GetConfiguration().GetAppNamespace(), name, func(ingress *extensions.Ingress) {
			for _, v := range this.ingressUFs {
				v(ingress)
			}
		})
		if err != nil {
			this.ctx.GetLog().Error(err, "Error during Ingress patching")
		}
	}
}

// ===

func (this *KubeClient) PatchDeployment(namespace, name string, updateFunc func(*apps.Deployment)) error {
	o, err := this.client.AppsV1().Deployments(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		return err
	}
	n := o.DeepCopy()
	updateFunc(n)
	patchData, err := createPatch(o, n, apps.Deployment{})
	if err != nil {
		return err
	}
	_, err = this.client.AppsV1beta1().Deployments(namespace).Patch(name, types.StrategicMergePatchType, patchData)
	return err
}

func (this *KubeClient) PatchService(namespace, name string, updateFunc ServiceUF) error {
	o, err := this.client.CoreV1().Services(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		return err
	}
	n := o.DeepCopy()
	updateFunc(n)
	patchData, err := createPatch(o, n, core.Service{})
	if err != nil {
		return err
	}
	_, err = this.client.CoreV1().Services(namespace).Patch(name, types.StrategicMergePatchType, patchData)
	return err
}

func (this *KubeClient) PatchIngress(namespace, name string, updateFunc func(*extensions.Ingress)) error {
	o, err := this.client.ExtensionsV1beta1().Ingresses(namespace).Get(name, meta.GetOptions{})
	if err != nil {
		return err
	}
	n := o.DeepCopy()
	updateFunc(n)
	patchData, err := createPatch(o, n, extensions.Ingress{})
	if err != nil {
		return err
	}
	_, err = this.client.ExtensionsV1beta1().Ingresses(namespace).Patch(name, types.StrategicMergePatchType, patchData)
	return err
}

// ===

func (this *KubePatcher) Execute() {
	if len(this.deploymentUFs) > 0 {
		this.patchDeployment()
	}
	if len(this.serviceUFs) > 0 {
		this.patchService()
	}
	if len(this.ingressUFs) > 0 {
		this.patchIngress()
	}
	// reset
	this.deploymentUFs = *new([]DeploymentUF)
	this.serviceUFs = *new([]ServiceUF)
	this.ingressUFs = *new([]IngressUF)
}
