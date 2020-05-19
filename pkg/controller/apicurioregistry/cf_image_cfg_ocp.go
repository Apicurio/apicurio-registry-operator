package apicurioregistry

import (
	registryv1alpha1 "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	ocp_apps "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ ControlFunction = &ImageConfigOCPCF{}

// This CF takes care of keeping the "image" section of the CRD applied.
type ImageConfigOCPCF struct {
	ctx *Context
}

func NewImageConfigOCPCF(ctx *Context) ControlFunction {
	return &ImageConfigOCPCF{ctx: ctx}
}

func (this *ImageConfigOCPCF) Describe() string {
	return "Image Configuration (OCP)"
}

func (this *ImageConfigOCPCF) Sense(spec *registryv1alpha1.ApicurioRegistry, request reconcile.Request) error {
	deployment, err := this.ctx.GetClients().OCP().GetCurrentDeployment()
	if err == nil {
		if c := deployment.Spec.Template.Spec.Containers; len(c) == 1 {
			this.ctx.GetConfiguration().SetConfig(CFG_STA_IMAGE, c[0].Image)
		} else {
			// TODO nuke the deployment?
			this.ctx.GetLog().Info("Warning: Unexpected contents of the Deployment resource '" + deployment.Name + "': More that one container")
		}
	} else {
		this.ctx.GetLog().Error(err, "error getting deployment")
	}
	return nil
}

func (this *ImageConfigOCPCF) Compare(spec *registryv1alpha1.ApicurioRegistry) (bool, error) {
	return this.ctx.GetConfiguration().GetConfig(CFG_STA_IMAGE) != this.ctx.GetConfiguration().GetImage(), nil
}

func (this *ImageConfigOCPCF) Respond(spec *registryv1alpha1.ApicurioRegistry) (bool, error) {
	this.ctx.GetPatchers().OCP().AddDeploymentPatch(func(deployment *ocp_apps.DeploymentConfig) {
		if c := deployment.Spec.Template.Spec.Containers; len(c) == 1 {
			c[0].Image = this.ctx.GetConfiguration().GetImage()
		} else {
			// TODO nuke the deployment?
			this.ctx.GetLog().Info("Warning: Unexpected contents of the Deployment resource '" + deployment.Name + "': More that one container")
		}
	})
	return true, nil
}
