package apicurioregistry

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	ocp_apps "github.com/openshift/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &DeploymentOCPCF{}

type DeploymentOCPCF struct {
	ctx *Context
}

func NewDeploymentOCPCF(ctx *Context) ControlFunction {

	err := ctx.GetController().Watch(&source.Kind{Type: &ocp_apps.DeploymentConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		fatal(ctx.GetLog(), err, "Error creating Deployment watch.")
	}

	return &DeploymentOCPCF{ctx: ctx}
}

func (this *DeploymentOCPCF) Describe() string {
	return "Deploymentconfiguration Creation (OCP)"
}

func (this *DeploymentOCPCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	// Try to check if there is an existing deployment resource
	deploymentName := this.ctx.GetConfiguration().GetConfig(CFG_STA_DEPLOYMENT_NAME)

	deployments, err := this.ctx.GetClients().OCP().GetDeployments(
		meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetConfiguration().GetAppName(),
		})
	if err != nil {
		return err
	}

	count := 0
	var lastDeployment *ocp_apps.DeploymentConfig = nil
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
		this.ctx.GetConfiguration().SetConfig(CFG_STA_DEPLOYMENT_NAME, lastDeployment.Name)
		return nil
	}
	// bad bad bad!
	this.ctx.GetLog().Info("Warning: Inconsistent Deployment state found.")
	this.ctx.GetConfiguration().ClearConfig(CFG_STA_DEPLOYMENT_NAME)
	for _, deployment := range deployments.Items {
		// nuke them...
		this.ctx.GetLog().Info("Warning: Deleting Deployment '" + deployment.Name + "'.")
		_ = this.ctx.GetClients().OCP().DeleteDeployment(deployment.Name, &meta.DeleteOptions{})
	}
	return nil
}

func (this *DeploymentOCPCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	return this.ctx.GetConfiguration().GetConfig(CFG_STA_DEPLOYMENT_NAME) == "", nil
}

func (this *DeploymentOCPCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	deployment := this.ctx.GetOCPFactory().CreateDeployment()

	if err := controllerutil.SetControllerReference(spec, deployment, this.ctx.GetScheme()); err != nil {
		this.ctx.GetLog().Error(err, "Cannot set controller reference.")
		return true, err
	}
	if err := this.ctx.GetNativeClient().Create(context.TODO(), deployment); err != nil { // Create runtime object is generic
		this.ctx.GetLog().Error(err, "Failed to create a new Deployment.")
		return true, err
	} else {
		this.ctx.GetConfiguration().SetConfig(CFG_STA_DEPLOYMENT_NAME, deployment.Name)
		this.ctx.GetLog().Info("New Deployment name is " + deployment.Name)
	}

	return true, nil
}
