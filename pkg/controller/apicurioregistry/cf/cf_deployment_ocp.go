package cf

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/status"
	ocp_apps "github.com/openshift/api/apps/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ loop.ControlFunction = &DeploymentOcpCF{}

type DeploymentOcpCF struct {
	ctx              *context.LoopContext
	svcResourceCache resources.ResourceCache
	svcClients       *client.Clients
	svcStatus        *status.Status
	svcOCPFactory    *factory.OCPFactory
	isCached         bool
	deployments      []ocp_apps.DeploymentConfig
	deploymentName   string
}

func NewDeploymentOcpCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {

	return &DeploymentOcpCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcClients:       services.Clients,
		svcStatus:        ctx.GetStatus(),
		svcOCPFactory:    services.OcpFactory,
		isCached:         false,
		deployments:      make([]ocp_apps.DeploymentConfig, 0),
		deploymentName:   resources.RC_EMPTY_NAME,
	}
}

func (this *DeploymentOcpCF) Describe() string {
	return "DeploymentOcpCF"
}

func (this *DeploymentOcpCF) Sense() {

	// Observation #1
	// Get cached Deployment
	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT_OCP)
	if deploymentExists {
		this.deploymentName = deploymentEntry.GetName().Str()
	} else {
		this.deploymentName = resources.RC_EMPTY_NAME
	}
	this.isCached = deploymentExists

	// Observation #2
	// Get deployment(s) we *should* track
	this.deployments = make([]ocp_apps.DeploymentConfig, 0)
	deployments, err := this.svcClients.OCP().GetDeployments(
		this.ctx.GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetAppName().Str(),
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

func (this *DeploymentOcpCF) Compare() bool {
	// Condition #1
	// If we already have a deployment cached, skip
	return !this.isCached
}

func (this *DeploymentOcpCF) Respond() {
	// Response #1
	// We already know about a deployment (name), and it is in the list
	if this.deploymentName != resources.RC_EMPTY_NAME {
		contains := false
		for _, val := range this.deployments {
			if val.Name == this.deploymentName {
				contains = true
				this.svcResourceCache.Set(resources.RC_KEY_DEPLOYMENT_OCP, resources.NewResourceCacheEntry(common.Name(val.Name), &val))
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
		this.svcResourceCache.Set(resources.RC_KEY_DEPLOYMENT_OCP, resources.NewResourceCacheEntry(common.Name(deployment.Name), &deployment))
	}
	// Response #3 (and #4)
	// If there is no deployment available (or there are more than 1), just create a new one
	if this.deploymentName == resources.RC_EMPTY_NAME && len(this.deployments) != 1 {
		deployment := this.svcOCPFactory.CreateDeployment()
		// leave the creation itself to patcher+creator so other CFs can update
		this.svcResourceCache.Set(resources.RC_KEY_DEPLOYMENT_OCP, resources.NewResourceCacheEntry(resources.RC_EMPTY_NAME, deployment))
	}
}

func (this *DeploymentOcpCF) Cleanup() bool {
	// Make sure the service is removed before we delete the deployment
	if _, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE); serviceExists {
		// Delete the service first
		return false
	}
	if deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT_OCP); deploymentExists {
		if err := this.svcClients.OCP().DeleteDeployment(deploymentEntry.GetValue().(*ocp_apps.DeploymentConfig), &meta.DeleteOptions{}); err != nil && !api_errors.IsNotFound(err) {
			this.ctx.GetLog().Error(err, "Could not delete deployment during cleanup")
			return false
		} else {
			this.svcResourceCache.Remove(resources.RC_KEY_DEPLOYMENT_OCP)
			this.ctx.GetLog().Info("Deployment has been deleted.")
		}
	}
	return true
}
