package apicurioregistry

import (
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// =====

type MonitoringClient struct {
	ctx             *Context
	client          *monclientv1.MonitoringV1Client
	discoveryClient *discovery.DiscoveryClient
}

func NewMonitoringClient(ctx *Context, config *rest.Config) *MonitoringClient {
	return &MonitoringClient{
		ctx:             ctx,
		client:          monclientv1.NewForConfigOrDie(config),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
	}
}

// ===
// ServiceMonitor

func (this *MonitoringClient) CreateServiceMonitor(namespace string, obj *monitoring.ServiceMonitor) (*monitoring.ServiceMonitor, error) {
	res, err := this.client.ServiceMonitors(namespace).Create(obj)
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetControllerReference(this.ctx.GetConfiguration().GetSpec(), res, this.ctx.GetScheme()); err != nil {
		panic("Could not set controller reference.")
	}
	res, err = this.UpdateServiceMonitor(namespace, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *MonitoringClient) GetServiceMonitor(namespace string, name string) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace).Get(name, v1.GetOptions{})
}

func (this *MonitoringClient) UpdateServiceMonitor(namespace string, obj *monitoring.ServiceMonitor) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace).Update(obj)
}

func (this *MonitoringClient) isServiceMonitorRegistered() (bool, error) {
	return k8sutil.ResourceExists(this.discoveryClient, "monitoring.coreos.com/v1", "ServiceMonitor")
}
