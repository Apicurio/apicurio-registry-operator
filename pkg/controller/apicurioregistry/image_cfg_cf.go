package apicurioregistry

import (
	registryv1alpha1 "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ ControlFunction = &ImageConfigCF{}

// This CF takes care of keeping the "image" section of the CRD applied.
type ImageConfigCF struct {
	ctx *Context
}

func NewImageConfigCF(ctx *Context) ControlFunction {
	return &ImageConfigCF{ctx: ctx}
}

func (this *ImageConfigCF) Describe() string {
	return "Image Configuration"
}

func (this *ImageConfigCF) Sense(spec *registryv1alpha1.ApicurioRegistry, request reconcile.Request) error {
	deployment, err := this.ctx.kubecl.GetDeployment()
	if err == nil {
		if c := deployment.Spec.Template.Spec.Containers; len(c) == 1 {
			this.ctx.configuration.SetConfig(CFG_STA_IMAGE, c[0].Image)
		} else {
			// TODO nuke the deployment?
			this.ctx.log.Info("Warning: Unexpected contents of the Deployment resource '" + deployment.Name + "': More that one container")
		}
	} else {
		this.ctx.log.Error(err, "error getting deployment")
	}
	return nil
}

func (this *ImageConfigCF) Compare(spec *registryv1alpha1.ApicurioRegistry) (bool, error) {
	return this.ctx.configuration.GetConfig(CFG_STA_IMAGE) != this.ctx.configuration.GetImage(), nil
}

func (this *ImageConfigCF) Respond(spec *registryv1alpha1.ApicurioRegistry) (bool, error) {
	this.ctx.patcher.AddDeploymentPatch(func(deployment *apps.Deployment) {
		if c := deployment.Spec.Template.Spec.Containers; len(c) == 1 {
			c[0].Image = this.ctx.configuration.GetImage()
		} else {
			// TODO nuke the deployment?
			log.Info("Warning: Unexpected contents of the Deployment resource '" + deployment.Name + "': More that one container")
		}
	})
	return true, nil
}
