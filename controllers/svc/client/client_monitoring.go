package client

import (
	ctx "context"
	"errors"

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
	spec := getSpec(this.ctx)
	if spec == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(spec, obj, this.ctx.GetScheme()); err != nil {
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

func (this *MonitoringClient) DeleteServiceMonitor(value *monitoring.ServiceMonitor, options *v1.DeleteOptions) error {
	return this.client.ServiceMonitors(value.Namespace).Delete(ctx.TODO(), value.Name, *options)
}
