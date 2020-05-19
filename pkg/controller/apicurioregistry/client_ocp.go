package apicurioregistry

import (
	"errors"
	ocp_apps "github.com/openshift/api/apps/v1"
	ocp_apps_client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type OCPClient struct {
	ctx           *Context
	ocpAppsClient *ocp_apps_client.AppsV1Client
}

func NewOCPClient(ctx *Context, clientConfig *rest.Config) *OCPClient {
	this := &OCPClient{
		ctx:           ctx,
		ocpAppsClient: ocp_apps_client.NewForConfigOrDie(clientConfig),
	}
	return this
}

func (this *OCPClient) GetCurrentDeployment() (*ocp_apps.DeploymentConfig, error) {
	if name := this.ctx.GetConfiguration().GetConfig(CFG_STA_DEPLOYMENT_NAME); name != "" {
		return this.GetDeployment(this.ctx.GetConfiguration().GetAppNamespace(), name, &meta.GetOptions{})
	}
	return nil, errors.New("No deployment name in status yet.")
}

func (this *OCPClient) GetDeployment(namespace string, name string, options *meta.GetOptions) (*ocp_apps.DeploymentConfig, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace).
		Get(name, *options)
}

func (this *OCPClient) DeleteDeployment(name string, options *meta.DeleteOptions) error {
	return this.ocpAppsClient.DeploymentConfigs(this.ctx.GetConfiguration().GetAppNamespace()).
		Delete(name, options)
}

func (this *OCPClient) GetDeployments(options meta.ListOptions) (*ocp_apps.DeploymentConfigList, error) {
	return this.ocpAppsClient.DeploymentConfigs(this.ctx.GetConfiguration().GetAppNamespace()).
		List(options)
}
