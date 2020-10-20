package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/status"
	apps "k8s.io/api/apps/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ loop.ControlFunction = &DeploymentCF{}

type DeploymentCF struct {
	ctx              loop.ControlLoopContext
	svcResourceCache resources.ResourceCache
	svcClients       *client.Clients
	svcStatus        *status.Status
	svcKubeFactory   *factory.KubeFactory
	isCached         bool
	deployments      []apps.Deployment
	deploymentName   string
}

func NewDeploymentCF(ctx loop.ControlLoopContext) loop.ControlFunction {

	err := ctx.GetController().Watch(&source.Kind{Type: &apps.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating watch.")
	}

	return &DeploymentCF{
		ctx:              ctx,
		svcResourceCache: ctx.RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache),
		svcClients:       ctx.RequireService(svc.SVC_CLIENTS).(*client.Clients),
		svcStatus:        ctx.RequireService(svc.SVC_STATUS).(*status.Status),
		svcKubeFactory:   ctx.RequireService(svc.SVC_KUBE_FACTORY).(*factory.KubeFactory),
		isCached:         false,
		deployments:      make([]apps.Deployment, 0),
		deploymentName:   resources.RC_EMPTY_NAME,
	}
}

func (this *DeploymentCF) Describe() string {
	return "DeploymentCF"
}

func (this *DeploymentCF) Sense() {

	// Observation #1
	// Get cached Deployment
	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)
	if deploymentExists {
		this.deploymentName = deploymentEntry.GetName()
	} else {
		this.deploymentName = resources.RC_EMPTY_NAME
	}
	this.isCached = deploymentExists

	// Observation #2
	// Get deployment(s) we *should* track
	this.deployments = make([]apps.Deployment, 0)
	deployments, err := this.svcClients.Kube().GetDeployments(
		this.ctx.GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetAppName(),
		})
	if err == nil {
		for _, deployment := range deployments.Items {
			if deployment.GetObjectMeta().GetDeletionTimestamp() == nil {
				this.deployments = append(this.deployments, deployment)
			}
		}
	}

	// Update the status
	this.svcStatus.SetConfig(status.CFG_STA_DEPLOYMENT_NAME, this.deploymentName)
}

func (this *DeploymentCF) Compare() bool {
	// Condition #1
	// If we already have a deployment cached, skip
	return !this.isCached
}

func (this *DeploymentCF) Respond() {
	// Response #1
	// We already know about a deployment (name), and it is in the list
	if this.deploymentName != resources.RC_EMPTY_NAME {
		contains := false
		for _, val := range this.deployments {
			if val.Name == this.deploymentName {
				contains = true
				this.svcResourceCache.Set(resources.RC_KEY_DEPLOYMENT, resources.NewResourceCacheEntry(val.Name, &val))
				break
			}
		}
		if !contains {
			this.deploymentName = resources.RC_EMPTY_NAME
		}
	}
	// Response #2
	// Can follow #1, but there must be a single deployment available
	if this.deploymentName == resources.RC_EMPTY_NAME && len(this.deployments) == 1 {
		deployment := this.deployments[0]
		this.deploymentName = deployment.Name
		this.svcResourceCache.Set(resources.RC_KEY_DEPLOYMENT, resources.NewResourceCacheEntry(deployment.Name, &deployment))
	}
	// Response #3 (and #4)
	// If there is no deployment available (or there are more than 1), just create a new one
	if this.deploymentName == resources.RC_EMPTY_NAME && len(this.deployments) != 1 {
		deployment := this.svcKubeFactory.CreateDeployment()
		// leave the creation itself to patcher+creator so other CFs can update
		this.svcResourceCache.Set(resources.RC_KEY_DEPLOYMENT, resources.NewResourceCacheEntry(resources.RC_EMPTY_NAME, deployment))
	}
}

func (this *DeploymentCF) Cleanup() bool {
	// Make sure the service is removed before we delete the deployment
	if _, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE); serviceExists {
		// Delete the service first
		return false
	}
	if deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); deploymentExists {
		if err := this.svcClients.Kube().DeleteDeployment(deploymentEntry.GetValue().(*apps.Deployment), &meta.DeleteOptions{});
			err != nil && !api_errors.IsNotFound(err) {
			this.ctx.GetLog().Error(err, "Could not delete deployment during cleanup.")
			return false
		} else {
			this.svcResourceCache.Remove(resources.RC_KEY_DEPLOYMENT)
			this.ctx.GetLog().Info("Deployment has been deleted.")
		}
	}
	return true
}
