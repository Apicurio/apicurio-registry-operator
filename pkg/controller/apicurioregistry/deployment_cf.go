package apicurioregistry

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &DeploymentCF{}

type DeploymentCF struct {
	ctx *Context
}

func NewDeploymentCF(ctx *Context) ControlFunction {

	err := ctx.c.Watch(&source.Kind{Type: &apps.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating Deployment watch.")
	}

	return &DeploymentCF{ctx: ctx}
}

func (this *DeploymentCF) Describe() string {
	return "Deployment Creation"
}

func (this *DeploymentCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	// Try to check if there is an existing deployment resource
	deploymentName := this.ctx.configuration.GetConfig(CFG_STA_DEPLOYMENT_NAME)

	deployments, err := this.ctx.kubecl.client.AppsV1().Deployments(this.ctx.configuration.GetSpecNamespace()).List(
		meta.ListOptions{
			LabelSelector: "app=" + this.ctx.configuration.GetSpecName(),
		})
	if err != nil {
		return err
	}

	count := 0
	var lastDeployment *apps.Deployment = nil
	for _, deployment := range deployments.Items {
		if deployment.GetObjectMeta().GetDeletionTimestamp() == nil {
			count++
			lastDeployment = &deployment
		}
	}

	if deploymentName == "" && count == 0 {
		// OK -> No dep. yet
		return nil
	}
	if deploymentName != "" && count == 1 && lastDeployment != nil && deploymentName == lastDeployment.Name {
		// OK -> dep exists
		return nil
	}
	if deploymentName == "" && count == 1 && lastDeployment != nil {
		// Also OK, but should not happen
		// save to status
		this.ctx.configuration.SetConfig(CFG_STA_DEPLOYMENT_NAME, lastDeployment.Name)
		return nil
	}
	// bad bad bad!
	this.ctx.log.Info("Warning: Inconsistent Deployment state found.")
	this.ctx.configuration.ClearConfig(CFG_STA_DEPLOYMENT_NAME)
	for _, deployment := range deployments.Items {
		// nuke them...
		this.ctx.log.Info("Warning: Deleting Deployment '" + deployment.Name + "'.")
		_ = this.ctx.kubecl.client.AppsV1().
			Deployments(this.ctx.configuration.GetSpecNamespace()).
			Delete(deployment.Name, &meta.DeleteOptions{})
	}
	return nil
}

func (this *DeploymentCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	return this.ctx.configuration.GetConfig(CFG_STA_DEPLOYMENT_NAME) == "", nil
}

func (this *DeploymentCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	deployment := this.ctx.factory.CreateDeployment()

	if err := controllerutil.SetControllerReference(spec, deployment, this.ctx.scheme); err != nil {
		log.Error(err, "Cannot set controller reference.")
		return true, err
	}
	if err := this.ctx.client.Create(context.TODO(), deployment); err != nil {
		log.Error(err, "Failed to create a new Deployment.")
		return true, err
	} else {
		this.ctx.configuration.SetConfig(CFG_STA_DEPLOYMENT_NAME, deployment.Name)
		log.Info("New Deployment name is " + deployment.Name)
	}

	return true, nil
}
