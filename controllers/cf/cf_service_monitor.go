package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
)

var _ loop.ControlFunction = &ServiceMonitorCF{}

type ServiceMonitorCF struct {
	ctx               context.LoopContext
	log               *zap.SugaredLogger
	svcResourceCache  resources.ResourceCache
	svcClients        *client.Clients
	monitoringFactory *factory.MonitoringFactory
	serviceMonitor    *monitoring.ServiceMonitor
	service           *core.Service
}

// TODO service monitor should be using resource cache
func NewServiceMonitorCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &ServiceMonitorCF{
		ctx:               ctx,
		svcResourceCache:  ctx.GetResourceCache(),
		svcClients:        ctx.GetClients(),
		monitoringFactory: services.GetMonitoringFactory(),
		serviceMonitor:    nil,
		service:           nil,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *ServiceMonitorCF) Describe() string {
	return "ServiceMonitorCF"
}

func (this *ServiceMonitorCF) Sense() {

	monitoringClient := this.svcClients.Monitoring()

	// Observation #2
	// Get Service
	serviceEntry, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
	if serviceExists {
		this.service = serviceEntry.GetValue().(*core.Service)
	}

	if serviceExists {
		// Observation #3
		// Get ServiceMonitor
		namespace := this.ctx.GetAppNamespace()
		name := this.ctx.GetAppName()
		serviceMonitor, err := monitoringClient.GetServiceMonitor(namespace, name)
		if err != nil {
			if !errors.IsNotFound(err) {
				this.log.Errorw("could not get ServiceMonitor", "error", err)
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
	return this.service != nil && this.serviceMonitor == nil
}

func (this *ServiceMonitorCF) Respond() {
	monitoringClient := this.svcClients.Monitoring()
	namespace := this.ctx.GetAppNamespace()
	serviceMonitor := this.monitoringFactory.NewServiceMonitor(this.service)

	// TODO FIX Should be using resource cache
	entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC)
	if !exists {
		return
	}
	_, err := monitoringClient.CreateServiceMonitor(entry.GetValue().(*ar.ApicurioRegistry), namespace, serviceMonitor)
	if err != nil {
		this.log.Errorw("could not create ServiceMonitor object", "error", err)
	}
}

func (this *ServiceMonitorCF) Cleanup() bool {
	// SM should not have any deletion dependencies
	monitoringClient := this.svcClients.Monitoring()
	namespace := this.ctx.GetAppNamespace()
	name := this.ctx.GetAppName()
	if serviceMonitor, err := monitoringClient.GetServiceMonitor(namespace, name); err == nil {
		if err := monitoringClient.DeleteServiceMonitor(serviceMonitor); err != nil && !api_errors.IsNotFound(err) /* Should not normally happen */ {
			this.log.Errorw("could not delete ServiceMonitor during cleanup", "error", err)
			return false
		} else {
			this.ctx.GetLog().Info("ServiceMonitor has been deleted.")
		}
	}

	return true
}
