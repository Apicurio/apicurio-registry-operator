package client

import (
	ctx "context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monclientv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// =====

type MonitoringClient struct {
	ctx             *context.LoopContext
	client          *monclientv1.MonitoringV1Client
	discoveryClient *discovery.DiscoveryClient
}

func NewMonitoringClient(ctx *context.LoopContext, config *rest.Config) *MonitoringClient {
	return &MonitoringClient{
		ctx:             ctx,
		client:          monclientv1.NewForConfigOrDie(config),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
	}
}

// ===
// ServiceMonitor

func (this *MonitoringClient) CreateServiceMonitor(namespace common.Namespace, obj *monitoring.ServiceMonitor) (*monitoring.ServiceMonitor, error) {
	if err := controllerutil.SetControllerReference(getSpec(this.ctx), obj, this.ctx.GetScheme()); err != nil {
		return nil, err
	}
	res, err := this.client.ServiceMonitors(namespace.Str()).Create(ctx.TODO(), obj, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *MonitoringClient) GetServiceMonitor(namespace common.Namespace, name common.Name) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace.Str()).Get(ctx.TODO(), name.Str(), v1.GetOptions{})
}

func (this *MonitoringClient) UpdateServiceMonitor(namespace common.Namespace, obj *monitoring.ServiceMonitor) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace.Str()).Update(ctx.TODO(), obj, v1.UpdateOptions{})
}

func (this *MonitoringClient) IsServiceMonitorRegistered() (bool, error) {
	return resourceExists(this.discoveryClient, "monitoring.coreos.com/v1", "ServiceMonitor")
}

func (this *MonitoringClient) DeleteServiceMonitor(value *monitoring.ServiceMonitor, options *v1.DeleteOptions) error {
	return this.client.ServiceMonitors(value.Namespace).Delete(ctx.TODO(), value.Name, *options)
}

// From k8sutil
func resourceExists(dc discovery.DiscoveryInterface, apiGroupVersion, kind string) (bool, error) {

	_, apiLists, err := dc.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == apiGroupVersion {
			for _, r := range apiList.APIResources {
				if r.Kind == kind {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
