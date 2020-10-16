package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/configuration"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	core "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ loop.ControlFunction = &ServiceCF{}

type ServiceCF struct {
	ctx            loop.ControlLoopContext
	isCached       bool
	services       []core.Service
	serviceName    string
	deploymentName string
}

func NewServiceCF(ctx loop.ControlLoopContext) loop.ControlFunction {

	err := ctx.GetController().Watch(&source.Kind{Type: &core.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating Service watch.")
	}

	return &ServiceCF{
		ctx:            ctx,
		isCached:       false,
		services:       make([]core.Service, 0),
		serviceName:    resources.RC_EMPTY_NAME,
		deploymentName: resources.RC_EMPTY_NAME,
	}
}

func (this *ServiceCF) Describe() string {
	return "ServiceCF"
}

func (this *ServiceCF) Sense() {

	// Observation #1
	// Get cached Service
	serviceEntry, serviceExists := this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Get(resources.RC_KEY_SERVICE)
	if serviceExists {
		this.serviceName = serviceEntry.GetName()
	} else {
		this.serviceName = resources.RC_EMPTY_NAME
	}
	this.isCached = serviceExists

	// Observation #2
	// Get service(s) we *should* track
	this.services = make([]core.Service, 0)
	services, err := this.ctx.RequireService(svc.SVC_CLIENTS).(*client.Clients).Kube().GetServices(
		this.ctx.GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetAppName(),
		})
	if err == nil {
		for _, service := range services.Items {
			if service.GetObjectMeta().GetDeletionTimestamp() == nil {
				this.services = append(this.services, service)
			}
		}
	}

	this.deploymentName = resources.RC_EMPTY_NAME

	// Observation #3
	// Is there a Deployment already? It must have been created (has a name)
	deploymentEntry, deploymentExists := this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Get(resources.RC_KEY_DEPLOYMENT)
	if deploymentExists {
		this.deploymentName = deploymentEntry.GetName()
	}

	// Observation #4
	// Same for OCP !!!
	deploymentEntry, deploymentExists = this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Get(resources.RC_KEY_DEPLOYMENT_OCP)
	if deploymentExists {
		this.deploymentName = deploymentEntry.GetName()
	}

	// Update the status
	this.ctx.RequireService(svc.SVC_CONFIGURATION).(*configuration.Configuration).SetConfig(configuration.CFG_STA_SERVICE_NAME, this.serviceName)
}

func (this *ServiceCF) Compare() bool {
	// Condition #1
	// If we already have a service cached, skip
	// Condition #2
	// The deployment has been created
	return !this.isCached && this.deploymentName != resources.RC_EMPTY_NAME
}

func (this *ServiceCF) Respond() {
	// Response #1
	// We already know about a service (name), and it is in the list
	if this.serviceName != resources.RC_EMPTY_NAME {
		contains := false
		for _, val := range this.services {
			if val.Name == this.serviceName {
				contains = true
				this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Set(resources.RC_KEY_SERVICE, resources.NewResourceCacheEntry(val.Name, &val))
				break
			}
		}
		if !contains {
			this.serviceName = resources.RC_EMPTY_NAME
		}
	}
	// Response #2
	// Can follow #1, but there must be a single service available
	if this.serviceName == resources.RC_EMPTY_NAME && len(this.services) == 1 {
		service := this.services[0]
		this.serviceName = service.Name
		this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Set(resources.RC_KEY_SERVICE, resources.NewResourceCacheEntry(service.Name, &service))
	}
	// Response #3 (and #4)
	// If there is no service available (or there are more than 1), just create a new one
	if this.serviceName == resources.RC_EMPTY_NAME && len(this.services) != 1 {
		service := this.ctx.RequireService(svc.SVC_KUBE_FACTORY).(*factory.KubeFactory).CreateService()
		// leave the creation itself to patcher+creator so other CFs can update
		this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Set(resources.RC_KEY_SERVICE, resources.NewResourceCacheEntry(resources.RC_EMPTY_NAME, service))
	}
}

func (this *ServiceCF) Cleanup() bool {
	// Make sure the ingress AND service monitor are removed before we delete the service
	// Ingress
	_, ingressExists := this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Get(resources.RC_KEY_INGRESS);
	// Service Monitor
	namespace := this.ctx.GetAppNamespace()
	name := this.ctx.GetAppName()
	_, serviceMonitorErr := this.ctx.RequireService(svc.SVC_CLIENTS).(*client.Clients).Monitoring().GetServiceMonitor(namespace, name);
	if ingressExists || (serviceMonitorErr != nil && !api_errors.IsNotFound(serviceMonitorErr) /* In case SM is not registered. */) {
		// Delete the ingress and SM first
		return false
	}
	if serviceEntry, serviceExists := this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Get(resources.RC_KEY_SERVICE); serviceExists {
		if err := this.ctx.RequireService(svc.SVC_CLIENTS).(*client.Clients).Kube().DeleteService(serviceEntry.GetValue().(*core.Service), &meta.DeleteOptions{});
			err != nil && !api_errors.IsNotFound(err) {
			this.ctx.GetLog().Error(err, "Could not delete service during cleanup")
			return false
		} else {
			this.ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Remove(resources.RC_KEY_SERVICE)
			this.ctx.GetLog().Info("Service has been deleted.")
		}
	}
	return true
}