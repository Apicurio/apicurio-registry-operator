package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ loop.ControlFunction = &ServiceMonitorCF{}

type ServiceMonitorCF struct {
	ctx                        loop.ControlLoopContext
	isServiceMonitorRegistered bool
	serviceMonitor             *monitoring.ServiceMonitor
	service                    *core.Service
}

// TODO service monitor should be using resource cache
func NewServiceMonitorCF(ctx loop.ControlLoopContext) loop.ControlFunction {

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

	monitoringClient := this.ctx.RequireService(svc.SVC_CLIENTS).(Clients).Monitoring()

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
	serviceEntry, serviceExists := this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(ResourceCache).Get(RC_KEY_SERVICE)
	if serviceExists {
		this.service = serviceEntry.GetValue().(*core.Service)
	}

	if isServiceMonitorRegistered && serviceExists {
		// Observation #3
		// Get ServiceMonitor
		namespace := this.ctx.GetAppNamespace()
		name := this.ctx.GetAppName()
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
	monitoringClient := this.ctx.RequireService(svc.SVC_CLIENTS).(Clients).Monitoring()
	monitoringFactory := NewMonitoringFactory(this.ctx)
	namespace := this.ctx.GetAppNamespace()
	serviceMonitor := monitoringFactory.NewServiceMonitor(this.service)

	_, err := monitoringClient.CreateServiceMonitor(namespace, serviceMonitor)
	if err != nil {
		log.Error(err, "Could not create ServiceMonitor object")
	}
}

func (this *ServiceMonitorCF) Cleanup() bool {
	// SM should not have any deletion dependencies
	monitoringClient := this.ctx.RequireService(svc.SVC_CLIENTS).(Clients).Monitoring()
	if isServiceMonitorRegistered, _ := monitoringClient.isServiceMonitorRegistered(); isServiceMonitorRegistered {
		namespace := this.ctx.GetAppNamespace()
		name := this.ctx.GetAppName()
		if serviceMonitor, err := monitoringClient.GetServiceMonitor(namespace, name); err == nil {
			if err := monitoringClient.DeleteServiceMonitor(serviceMonitor, &meta.DeleteOptions{});
				err != nil && !api_errors.IsNotFound(err) /* Should not normally happen */ {
				this.ctx.GetLog().Error(err, "Could not delete ServiceMonitor during cleanup.")
				return false
			} else {
				this.ctx.GetLog().Info("ServiceMonitor has been deleted.")
			}
		}
	}
	return true
}
