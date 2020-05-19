package apicurioregistry

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"k8s.io/api/extensions/v1beta1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &IngressCF{}

type IngressCF struct {
	ctx *Context
}

func NewIngressCF(ctx *Context) ControlFunction {

	err := ctx.GetController().Watch(&source.Kind{Type: &v1beta1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating Ingress watch.")
	}

	return &IngressCF{ctx: ctx}
}

func (this *IngressCF) Describe() string {
	return "Ingress Creation"
}

func (this *IngressCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	ingressName := this.ctx.GetConfiguration().GetConfig(CFG_STA_INGRESS_NAME)

	ingresses, err := this.ctx.GetClients().Kube().GetRawClient().ExtensionsV1beta1().Ingresses(this.ctx.GetConfiguration().GetAppNamespace()).List(
		meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetConfiguration().GetAppName(),
		})
	if err != nil {
		return err
	}

	count := 0
	var lastIngress *extensions.Ingress = nil
	for _, ingress := range ingresses.Items {
		if ingress.GetObjectMeta().GetDeletionTimestamp() == nil {
			count++
			lastIngress = &ingress
		}
	}

	if ingressName == "" && count == 0 {
		// OK -> No dep. yet
		return nil
	}
	if ingressName != "" && count == 1 && lastIngress != nil && ingressName == lastIngress.Name {
		// OK -> dep exists
		return nil
	}
	if ingressName == "" && count == 1 && lastIngress != nil {
		// Also OK, but should not happen
		// save to status
		this.ctx.GetConfiguration().SetConfig(CFG_STA_INGRESS_NAME, lastIngress.Name)
		return nil
	}
	// bad bad bad!
	this.ctx.GetLog().Info("Warning: Inconsistent Ingress state found.")
	this.ctx.GetConfiguration().ClearConfig(CFG_STA_INGRESS_NAME)
	for _, ingress := range ingresses.Items {
		// nuke them...
		this.ctx.GetLog().Info("Warning: Deleting Ingress '" + ingress.Name + "'.")
		_ = this.ctx.GetClients().Kube().GetRawClient().ExtensionsV1beta1().
			Ingresses(this.ctx.GetConfiguration().GetAppNamespace()).
			Delete(ingress.Name, &meta.DeleteOptions{})
	}
	return nil
}

func (this *IngressCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	// Do not create ingress if "route" is not defined
	return this.ctx.GetConfiguration().GetConfig(CFG_DEP_ROUTE) != "" &&
		this.ctx.GetConfiguration().GetConfig(CFG_STA_INGRESS_NAME) == "" &&
		this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME) != "", nil
}

func (this *IngressCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	serviceName := this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME)
	if serviceName == "" {
		return true, nil
	}
	ingress := this.ctx.GetKubeFactory().CreateIngress(serviceName)

	if err := controllerutil.SetControllerReference(spec, ingress, this.ctx.GetScheme()); err != nil {
		this.ctx.GetLog().Error(err, "Cannot set controller reference.")
		return true, err
	}
	if err := this.ctx.GetNativeClient().Create(context.TODO(), ingress); err != nil {
		this.ctx.GetLog().Error(err, "Failed to create a new Ingress.")
		return true, err
	} else {
		this.ctx.GetConfiguration().SetConfig(CFG_STA_INGRESS_NAME, ingress.Name)
		this.ctx.GetLog().Info("New Ingress name is " + ingress.Name)
	}

	return true, nil
}
