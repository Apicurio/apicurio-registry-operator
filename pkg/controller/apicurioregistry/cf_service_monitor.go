package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &ServiceMonitorCF{}

type ServiceMonitorCF struct {
	ctx                        *Context
	isServiceMonitorRegistered bool
	serviceMonitor             *monitoring.ServiceMonitor
	service                    *core.Service
}

func NewServiceMonitorCF(ctx *Context) ControlFunction {

	return &ServiceMonitorCF{
		ctx:                        ctx,
		isServiceMonitorRegistered: false,
		serviceMonitor:             nil,
		service:                    nil,
	}
}

func (this *ServiceMonitorCF) Describe() string {
	return "ServiceMonitorCF"
}

func (this *ServiceMonitorCF) Sense() {

	err := this.ctx.GetController().Watch(&source.Kind{Type: &monitoring.ServiceMonitor{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		this.ctx.GetLog().WithValues("type", "Warning", "reason", err.Error()).
			Info("Could not create ServiceMonitor watch.")
		return
	}

	monitoringClient := this.ctx.GetClients().Monitoring()

	// Observation #1
	// Is ServiceMonitor registered?
	isServiceMonitorRegistered, err := monitoringClient.isServiceMonitorRegistered()
	if err != nil {
		log.Error(err, "Could not check ServiceMonitor is registered")
		return
	}
	if !isServiceMonitorRegistered {
		log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects")
	}
	this.isServiceMonitorRegistered = isServiceMonitorRegistered

	// Observation #2
	// Get Service
	serviceEntry, serviceExists := this.ctx.GetResourceCache().Get(RC_KEY_SERVICE)
	if serviceExists {
		this.service = serviceEntry.GetValue().(*core.Service)
	}

	if isServiceMonitorRegistered && serviceExists {
		// Observation #3
		// Get ServiceMonitor
		namespace := this.ctx.configuration.GetAppNamespace()
		name := this.ctx.configuration.GetAppName()
		serviceMonitor, err := monitoringClient.GetServiceMonitor(namespace, name)
		if err != nil {
			if !errors.IsNotFound(err) {
				log.Error(err, "Could not get ServiceMonitor")
			}
			return
		}
		this.serviceMonitor = serviceMonitor
	}
}

func (this *ServiceMonitorCF) Compare() bool {
	// Condition #1
	// ServiceMonitor is registered
	// Condition #2
	// Service has been created
	// Condition #3
	// ServiceMonitor has not been created
	return this.isServiceMonitorRegistered && this.service != nil && this.serviceMonitor == nil
}

func (this *ServiceMonitorCF) Respond() {
	monitoringClient := this.ctx.GetClients().Monitoring()
	monitoringFactory := NewMonitoringFactory(this.ctx)
	namespace := this.ctx.configuration.GetAppNamespace()
	serviceMonitor := monitoringFactory.NewServiceMonitor(this.service)

	_, err := monitoringClient.CreateServiceMonitor(namespace, serviceMonitor)
	if err != nil {
		log.Error(err, "Could not create ServiceMonitor object")
	}
}
