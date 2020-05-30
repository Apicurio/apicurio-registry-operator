package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &DeploymentCF{}

type DeploymentCF struct {
	ctx            *Context
	isCached       bool
	deployments    []apps.Deployment
	deploymentName string
}

func NewDeploymentCF(ctx *Context) ControlFunction {

	err := ctx.GetController().Watch(&source.Kind{Type: &apps.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating watch.")
	}

	return &DeploymentCF{
		ctx:            ctx,
		isCached:       false,
		deployments:    make([]apps.Deployment, 0),
		deploymentName: RC_EMPTY_NAME,
	}
}

func (this *DeploymentCF) Describe() string {
	return "DeploymentCF"
}

func (this *DeploymentCF) Sense() {

	// Observation #1
	// Get cached Deployment
	deploymentEntry, deploymentExists := this.ctx.GetResourceCache().Get(RC_KEY_DEPLOYMENT)
	if deploymentExists {
		this.deploymentName = deploymentEntry.GetName()
	} else {
		this.deploymentName = RC_EMPTY_NAME
	}
	this.isCached = deploymentExists

	// Observation #2
	// Get deployment(s) we *should* track
	this.deployments = make([]apps.Deployment, 0)
	deployments, err := this.ctx.GetClients().Kube().GetDeployments(
		this.ctx.GetConfiguration().GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetConfiguration().GetAppName(),
		})
	if err == nil {
		for _, deployment := range deployments.Items {
			if deployment.GetObjectMeta().GetDeletionTimestamp() == nil {
				this.deployments = append(this.deployments, deployment)
			}
		}
	}

	// Update the status
	this.ctx.GetConfiguration().SetConfig(CFG_STA_DEPLOYMENT_NAME, this.deploymentName)
}

func (this *DeploymentCF) Compare() bool {
	// Condition #1
	// If we already have a deployment cached, skip
	return !this.isCached
}

func (this *DeploymentCF) Respond() {
	// Response #1
	// We already know about a deployment (name), and it is in the list
	if this.deploymentName != RC_EMPTY_NAME {
		contains := false
		for _, val := range this.deployments {
			if val.Name == this.deploymentName {
				contains = true
				this.ctx.GetResourceCache().Set(RC_KEY_DEPLOYMENT, NewResourceCacheEntry(val.Name, &val))
				break
			}
		}
		if !contains {
			this.deploymentName = RC_EMPTY_NAME
		}
	}
	// Response #2
	// Can follow #1, but there must be a single deployment available
	if this.deploymentName == RC_EMPTY_NAME && len(this.deployments) == 1 {
		deployment := this.deployments[0]
		this.deploymentName = deployment.Name
		this.ctx.GetResourceCache().Set(RC_KEY_DEPLOYMENT, NewResourceCacheEntry(deployment.Name, &deployment))
	}
	// Response #3 (and #4)
	// If there is no deployment available (or there are more than 1), just create a new one
	if this.deploymentName == RC_EMPTY_NAME && len(this.deployments) != 1 {
		deployment := this.ctx.GetKubeFactory().CreateDeployment()
		// leave the creation itself to patcher+creator so other CFs can update
		this.ctx.GetResourceCache().Set(RC_KEY_DEPLOYMENT, NewResourceCacheEntry(RC_EMPTY_NAME, deployment))
	}
}
