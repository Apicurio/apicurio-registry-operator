package apicurioregistry

import (
	"context"
	ar "github.com/apicurio/apicurio-operators/apicurio-registry/pkg/apis/apicur/v1alpha1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &ServiceCF{}

type ServiceCF struct {
	ctx *Context
}

func NewServiceCF(ctx *Context) ControlFunction {

	err := ctx.c.Watch(&source.Kind{Type: &core.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating Service watch!")
	}

	return &ServiceCF{ctx: ctx}
}

func (this *ServiceCF) Describe() string {
	return "Service Creation"
}

func (this *ServiceCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {

	serviceName := this.ctx.configuration.GetConfig(CFG_STA_SERVICE_NAME)

	services, err := this.ctx.kubecl.client.CoreV1().Services(this.ctx.configuration.GetSpecNamespace()).List(
		meta.ListOptions{
			LabelSelector: "app=" + this.ctx.configuration.GetSpecName(),
		})
	if err != nil {
		return err
	}

	count := 0
	var lastService *core.Service = nil
	for _, service := range services.Items {
		if service.GetObjectMeta().GetDeletionTimestamp() == nil {
			count++
			lastService = &service
		}
	}

	if serviceName == "" && count == 0 {
		// OK -> No svc. yet
		return nil
	}
	if serviceName != "" && count == 1 && lastService != nil && serviceName == lastService.Name {
		// OK -> svc. exists
		return nil
	}
	if serviceName == "" && count == 1 && lastService != nil {
		// Also OK, but should not happen
		// save to status
		this.ctx.configuration.SetConfig(CFG_STA_SERVICE_NAME, lastService.Name)
		return nil
	}
	// bad bad bad!
	this.ctx.log.Info("Warning: Inconsistent Service state found.")
	this.ctx.configuration.ClearConfig(CFG_STA_SERVICE_NAME)
	for _, service := range services.Items {
		// nuke them...
		this.ctx.log.Info("Warning: Deleting Service '" + service.Name + "'.")
		_ = this.ctx.kubecl.client.AppsV1().
			Deployments(this.ctx.configuration.GetSpecNamespace()).
			Delete(service.Name, &meta.DeleteOptions{})
	}
	return nil
}

func (this *ServiceCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	return this.ctx.configuration.GetConfig(CFG_STA_SERVICE_NAME) == "", nil
}

func (this *ServiceCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	service := this.ctx.factory.CreateService()

	if err := controllerutil.SetControllerReference(spec, service, this.ctx.scheme); err != nil {
		log.Error(err, "Cannot set controller reference.")
		return true, err
	}
	if err := this.ctx.client.Create(context.TODO(), service); err != nil {
		log.Error(err, "Failed to create a new Service.")
		return true, err
	} else {
		this.ctx.configuration.SetConfig(CFG_STA_SERVICE_NAME, service.Name)
		log.Info("New Service name is " + service.Name)
	}

	return true, nil
}
