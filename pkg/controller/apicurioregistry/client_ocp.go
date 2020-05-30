package apicurioregistry

import (
	ocp_apps "github.com/openshift/api/apps/v1"
	ocp_apps_client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// ===
// Deployment

func (this *OCPClient) CreateDeployment(namespace string, value *ocp_apps.DeploymentConfig) (*ocp_apps.DeploymentConfig, error) {
	res, err := this.ocpAppsClient.DeploymentConfigs(namespace).
		Create(value)
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetControllerReference(this.ctx.GetConfiguration().GetSpec(), res, this.ctx.GetScheme()); err != nil {
		panic("Could not set controller reference.")
	}
	res, err = this.UpdateDeployment(namespace, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *OCPClient) GetDeployment(namespace string, name string, options *meta.GetOptions) (*ocp_apps.DeploymentConfig, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace).
		Get(name, *options)
}

func (this *OCPClient) UpdateDeployment(namespace string, value *ocp_apps.DeploymentConfig) (*ocp_apps.DeploymentConfig, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace).
		Update(value)
}

func (this *OCPClient) PatchDeployment(namespace, name string, patchData []byte) (*ocp_apps.DeploymentConfig, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace).
		Patch(name, types.StrategicMergePatchType, patchData)
}

func (this *OCPClient) DeleteDeployment(namespace string, name string, options *meta.DeleteOptions) error {
	return this.ocpAppsClient.DeploymentConfigs(namespace).
		Delete(name, options)
}

func (this *OCPClient) GetDeployments(namespace string, options *meta.ListOptions) (*ocp_apps.DeploymentConfigList, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace).
		List(*options)
}
