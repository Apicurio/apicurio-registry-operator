package client

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
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
	res, err := this.client.ServiceMonitors(namespace.Str()).Create(obj)
	if err != nil {
		return nil, err
	}
	if err := controllerutil.SetControllerReference(getSpec(this.ctx), res, this.ctx.GetScheme()); err != nil {
		panic("Could not set controller reference.")
	}
	res, err = this.UpdateServiceMonitor(namespace, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *MonitoringClient) GetServiceMonitor(namespace common.Namespace, name common.Name) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace.Str()).Get(name.Str(), v1.GetOptions{})
}

func (this *MonitoringClient) UpdateServiceMonitor(namespace common.Namespace, obj *monitoring.ServiceMonitor) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace.Str()).Update(obj)
}

func (this *MonitoringClient) IsServiceMonitorRegistered() (bool, error) {
	return k8sutil.ResourceExists(this.discoveryClient, "monitoring.coreos.com/v1", "ServiceMonitor")
}

func (this *MonitoringClient) DeleteServiceMonitor(value *monitoring.ServiceMonitor, options *v1.DeleteOptions) error {
	return this.client.ServiceMonitors(value.Namespace).Delete(value.Name, options)
}
