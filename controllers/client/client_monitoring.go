package client

import (
	ctx "context"
	"errors"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/go-logr/logr"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	monclientv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// =====

type MonitoringClient struct {
	//ctx             context.LoopContext
	log    logr.Logger
	client *monclientv1.MonitoringV1Client
	scheme *runtime.Scheme

	//discoveryClient *discovery.DiscoveryClient
}

func NewMonitoringClient(log logr.Logger, scheme *runtime.Scheme, config *rest.Config) *MonitoringClient {
	return &MonitoringClient{
		log:    log,
		client: monclientv1.NewForConfigOrDie(config),
		scheme: scheme,
		//discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
	}
}

// ===
// ServiceMonitor

func (this *MonitoringClient) CreateServiceMonitor(owner meta.Object, namespace common.Namespace, obj *monitoring.ServiceMonitor) (*monitoring.ServiceMonitor, error) {
	if owner == nil {
		return nil, errors.New("Could not find ApicurioRegistry. Retrying.")
	}
	if err := controllerutil.SetControllerReference(owner, obj, this.scheme); err != nil {
		return nil, err
	}
	res, err := this.client.ServiceMonitors(namespace.Str()).Create(ctx.TODO(), obj, meta.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *MonitoringClient) GetServiceMonitor(namespace common.Namespace, name common.Name) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace.Str()).Get(ctx.TODO(), name.Str(), meta.GetOptions{})
}

func (this *MonitoringClient) UpdateServiceMonitor(namespace common.Namespace, obj *monitoring.ServiceMonitor) (*monitoring.ServiceMonitor, error) {
	return this.client.ServiceMonitors(namespace.Str()).Update(ctx.TODO(), obj, meta.UpdateOptions{})
}

func (this *MonitoringClient) DeleteServiceMonitor(value *monitoring.ServiceMonitor) error {
	return this.client.ServiceMonitors(value.Namespace).Delete(ctx.TODO(), value.Name, meta.DeleteOptions{})
}
