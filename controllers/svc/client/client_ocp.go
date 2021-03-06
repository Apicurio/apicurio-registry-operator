package client

import (
	ctx "context"
	"errors"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	ocp_apps "github.com/openshift/api/apps/v1"
	ocp_route "github.com/openshift/api/route/v1"
	ocp_apps_client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	ocp_route_client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type OCPClient struct {
	ctx            *context.LoopContext
	ocpAppsClient  *ocp_apps_client.AppsV1Client
	ocpRouteClient *ocp_route_client.RouteV1Client
}

func NewOCPClient(ctx *context.LoopContext, clientConfig *rest.Config) *OCPClient {
	this := &OCPClient{
		ctx:            ctx,
		ocpAppsClient:  ocp_apps_client.NewForConfigOrDie(clientConfig),
		ocpRouteClient: ocp_route_client.NewForConfigOrDie(clientConfig),
	}
	return this
}

// ===
// Deployment

func (this *OCPClient) CreateDeployment(namespace common.Namespace, value *ocp_apps.DeploymentConfig) (*ocp_apps.DeploymentConfig, error) {
	spec := getSpec(this.ctx)
	if spec == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(spec, value, this.ctx.GetScheme()); err != nil {
		return nil, err
	}
	res, err := this.ocpAppsClient.DeploymentConfigs(namespace.Str()).Create(ctx.TODO(), value, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *OCPClient) GetDeployment(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*ocp_apps.DeploymentConfig, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace.Str()).
		Get(ctx.TODO(), name.Str(), *options)
}

func (this *OCPClient) UpdateDeployment(namespace common.Namespace, value *ocp_apps.DeploymentConfig) (*ocp_apps.DeploymentConfig, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace.Str()).
		Update(ctx.TODO(), value, meta.UpdateOptions{})
}

func (this *OCPClient) PatchDeployment(namespace common.Namespace, name common.Name, patchData []byte) (*ocp_apps.DeploymentConfig, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace.Str()).
		Patch(ctx.TODO(), name.Str(), types.StrategicMergePatchType, patchData, meta.PatchOptions{})
}

func (this *OCPClient) DeleteDeployment(value *ocp_apps.DeploymentConfig, options *meta.DeleteOptions) error {
	return this.ocpAppsClient.DeploymentConfigs(value.Namespace).
		Delete(ctx.TODO(), value.Name, *options)
}

func (this *OCPClient) GetDeployments(namespace common.Namespace, options *meta.ListOptions) (*ocp_apps.DeploymentConfigList, error) {
	return this.ocpAppsClient.DeploymentConfigs(namespace.Str()).
		List(ctx.TODO(), *options)
}

// ======
// Route

func (this *OCPClient) GetRoute(namespace common.Namespace, name common.Name, options *meta.GetOptions) (*ocp_route.Route, error) {
	return this.ocpRouteClient.Routes(namespace.Str()).
		Get(ctx.TODO(), name.Str(), *options)
}

func (this *OCPClient) GetRoutes(namespace common.Namespace, options *meta.ListOptions) (*ocp_route.RouteList, error) {
	return this.ocpRouteClient.Routes(namespace.Str()).
		List(ctx.TODO(), *options)
}
