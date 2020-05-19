package apicurioregistry

import (
	"errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// =====

type KubeClient struct {
	ctx    *Context
	client kubernetes.Interface
}

func NewKubeClient(ctx *Context, config *rest.Config) *KubeClient {
	return &KubeClient{
		client: kubernetes.NewForConfigOrDie(config),
		ctx:    ctx,
	}
}

// =====

func (this *KubeClient) GetCurrentDeployment() (*apps.Deployment, error) {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_DEPLOYMENT_NAME); name != "" {
		deployment, err := this.client.AppsV1().Deployments(this.ctx.GetConfiguration().GetAppNamespace()).Get(name, meta.GetOptions{})
		return deployment, err
	}
	return nil, errors.New("No deployment name in status yet.")
}

func (this *KubeClient) GetCurrentService() (*core.Service, error) {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME); name != "" {
		service, err := this.client.CoreV1().Services(this.ctx.GetConfiguration().GetAppNamespace()).Get(name, meta.GetOptions{})
		return service, err
	}
	return nil, errors.New("No service name in status yet.")
}

func (this *KubeClient) GetCurrentIngress() (*extensions.Ingress, error) {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_INGRESS_NAME); name != "" {
		ingress, err := this.client.ExtensionsV1beta1().Ingresses(this.ctx.GetConfiguration().GetAppNamespace()).Get(name, meta.GetOptions{})
		return ingress, err
	}
	return nil, errors.New("No ingress name in status yet.")
}

func (this *KubeClient) GetRawClient() kubernetes.Interface {
	return this.client
}
