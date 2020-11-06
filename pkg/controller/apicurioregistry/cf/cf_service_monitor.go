package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ loop.ControlFunction = &ServiceMonitorCF{}

type ServiceMonitorCF struct {
	ctx                        *context.LoopContext
	svcResourceCache           resources.ResourceCache
	svcClients                 *client.Clients
	monitoringFactory          *factory.MonitoringFactory
	isServiceMonitorRegistered bool
	serviceMonitor             *monitoring.ServiceMonitor
	service                    *core.Service
}

// TODO service monitor should be using resource cache
func NewServiceMonitorCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {

	return &ServiceMonitorCF{
		ctx:                        ctx,
		svcResourceCache:           ctx.GetResourceCache(),
		svcClients:                 services.Clients,
		monitoringFactory:          services.MonitoringFactory,
		isServiceMonitorRegistered: false,
		serviceMonitor:             nil,
		service:                    nil,
	}
}

func (this *ServiceMonitorCF) Describe() string {
	return "ServiceMonitorCF"
}

func (this *ServiceMonitorCF) Sense() {

	monitoringClient := this.svcClients.Monitoring()

	// Observation #1
	// Is ServiceMonitor registered?
	isServiceMonitorRegistered, err := monitoringClient.IsServiceMonitorRegistered()
	if err != nil {
		this.ctx.GetLog().Error(err, "Could not check ServiceMonitor is registered")
		return
	}
	if !isServiceMonitorRegistered {
		this.ctx.GetLog().Info("Install prometheus-operator in your cluster to create ServiceMonitor objects")
	}
	this.isServiceMonitorRegistered = isServiceMonitorRegistered

	// Observation #2
	// Get Service
	serviceEntry, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
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
				this.ctx.GetLog().Error(err, "Could not get ServiceMonitor")
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
	monitoringClient := this.svcClients.Monitoring()
	namespace := this.ctx.GetAppNamespace()
	serviceMonitor := this.monitoringFactory.NewServiceMonitor(this.service)

	_, err := monitoringClient.CreateServiceMonitor(namespace, serviceMonitor)
	if err != nil {
		this.ctx.GetLog().Error(err, "Could not create ServiceMonitor object")
	}
}

func (this *ServiceMonitorCF) Cleanup() bool {
	// SM should not have any deletion dependencies
	monitoringClient := this.svcClients.Monitoring()
	if isServiceMonitorRegistered, _ := monitoringClient.IsServiceMonitorRegistered(); isServiceMonitorRegistered {
		namespace := this.ctx.GetAppNamespace()
		name := this.ctx.GetAppName()
		if serviceMonitor, err := monitoringClient.GetServiceMonitor(namespace, name); err == nil {
			if err := monitoringClient.DeleteServiceMonitor(serviceMonitor, &meta.DeleteOptions{}); err != nil && !api_errors.IsNotFound(err) /* Should not normally happen */ {
				this.ctx.GetLog().Error(err, "Could not delete ServiceMonitor during cleanup.")
				return false
			} else {
				this.ctx.GetLog().Info("ServiceMonitor has been deleted.")
			}
		}
	}
	return true
}
