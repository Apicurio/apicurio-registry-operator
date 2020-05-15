package apicurioregistry

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
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

	err := ctx.GetController().Watch(&source.Kind{Type: &core.Service{}}, &handler.EnqueueRequestForOwner{
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

	serviceName := this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME)

	services, err := this.ctx.GetKubeCl().GetClient().CoreV1().Services(this.ctx.GetConfiguration().GetSpecNamespace()).List(
		meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetConfiguration().GetSpecName(),
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
		this.ctx.GetConfiguration().SetConfig(CFG_STA_SERVICE_NAME, lastService.Name)
		return nil
	}
	// bad bad bad!
	this.ctx.GetLog().Info("Warning: Inconsistent Service state found.")
	this.ctx.GetConfiguration().ClearConfig(CFG_STA_SERVICE_NAME)
	for _, service := range services.Items {
		// nuke them...
		this.ctx.GetLog().Info("Warning: Deleting Service '" + service.Name + "'.")
		_ = this.ctx.GetKubeCl().GetClient().AppsV1().
			Deployments(this.ctx.GetConfiguration().GetSpecNamespace()).
			Delete(service.Name, &meta.DeleteOptions{})
	}
	return nil
}

func (this *ServiceCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	return this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME) == "", nil
}

func (this *ServiceCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	service := this.ctx.GetFactory().CreateService()

	if err := controllerutil.SetControllerReference(spec, service, this.ctx.GetScheme()); err != nil {
		this.ctx.GetLog().Error(err, "Cannot set controller reference.")
		return true, err
	}
	if err := this.ctx.GetClient().Create(context.TODO(), service); err != nil {
		this.ctx.GetLog().Error(err, "Failed to create a new Service.")
		return true, err
	} else {
		this.ctx.GetConfiguration().SetConfig(CFG_STA_SERVICE_NAME, service.Name)
		this.ctx.GetLog().Info("New Service name is " + service.Name)
	}

	return true, nil
}
