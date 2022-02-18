package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	networking "k8s.io/api/networking/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ loop.ControlFunction = &NetworkPolicyCF{}

type NetworkPolicyCF struct {
	ctx               *context.LoopContext
	svcResourceCache  resources.ResourceCache
	svcClients        *client.Clients
	svcStatus         *status.Status
	svcKubeFactory    *factory.KubeFactory
	isCached          bool
	networkPolicies   []networking.NetworkPolicy
	networkPolicyName string
	serviceName       string
	targetHostIsEmpty bool
}

func NewNetworkPolicyCF(ctx *context.LoopContext, services *services.LoopServices) loop.ControlFunction {

	return &NetworkPolicyCF{
		ctx:               ctx,
		svcResourceCache:  ctx.GetResourceCache(),
		svcClients:        services.GetClients(),
		svcStatus:         services.GetStatus(),
		svcKubeFactory:    services.GetKubeFactory(),
		isCached:          false,
		networkPolicies:   make([]networking.NetworkPolicy, 0),
		networkPolicyName: resources.RC_EMPTY_NAME,
		serviceName:       resources.RC_EMPTY_NAME,
		targetHostIsEmpty: true,
	}
}

func (this *NetworkPolicyCF) Describe() string {
	return "NetworkPolicyCF"
}

func (this *NetworkPolicyCF) Sense() {

	// Observation #1
	// Get cached Network Policy
	networkPolicyEntry, networkPolicyExists := this.svcResourceCache.Get(resources.RC_KEY_NETWORK_POLICY)
	if networkPolicyExists {
		this.networkPolicyName = networkPolicyEntry.GetName().Str()
	} else {
		this.networkPolicyName = resources.RC_EMPTY_NAME
	}
	this.isCached = networkPolicyExists

	// Observation #2
	// Get networkPolicy(s) we *should* track
	this.networkPolicies = make([]networking.NetworkPolicy, 0)
	networkPolicies, err := this.svcClients.Kube().GetNetworkPolicies(
		this.ctx.GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetAppName().Str(),
		})
	if err == nil {
		for _, networkPolicy := range networkPolicies.Items {
			if networkPolicy.GetObjectMeta().GetDeletionTimestamp() == nil {
				this.networkPolicies = append(this.networkPolicies, networkPolicy)
			}
		}
	}

	// Observation #3
	// Is there a Service already? It must have been created (has a name)
	serviceEntry, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
	if serviceExists {
		this.serviceName = serviceEntry.GetName().Str()
	} else {
		this.serviceName = resources.RC_EMPTY_NAME
	}

	// Observation #4
	// See if the host in the config spec is not empty
	this.targetHostIsEmpty = true
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.targetHostIsEmpty = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host == ""
	}

	// Update the status
	this.svcStatus.SetConfig(status.CFG_STA_NETWORK_POLICY_NAME, this.networkPolicyName)
}

func (this *NetworkPolicyCF) Compare() bool {
	// Condition #1
	// If we already have a networkPolicy cached, skip
	// Condition #2
	// The service has been created
	// Condition #3
	// We will create a new networkPolicy only if the host is not empty
	return !this.isCached &&
		this.serviceName != resources.RC_EMPTY_NAME &&
		!this.targetHostIsEmpty
}

func (this *NetworkPolicyCF) Respond() {
	// Response #1
	// We already know about a networkPolicy (name), and it is in the list
	if this.networkPolicyName != resources.RC_EMPTY_NAME {
		contains := false
		for _, val := range this.networkPolicies {
			if val.Name == this.networkPolicyName {
				contains = true
				this.svcResourceCache.Set(resources.RC_KEY_NETWORK_POLICY, resources.NewResourceCacheEntry(common.Name(val.Name), &val))
				break
			}
		}
		if !contains {
			this.networkPolicyName = resources.RC_EMPTY_NAME
		}
	}
	// Response #2
	// Can follow #1, but there must be a single networkPolicy available
	if this.networkPolicyName == resources.RC_EMPTY_NAME && len(this.networkPolicies) == 1 {
		networkPolicy := this.networkPolicies[0]
		this.networkPolicyName = networkPolicy.Name
		this.svcResourceCache.Set(resources.RC_KEY_NETWORK_POLICY, resources.NewResourceCacheEntry(common.Name(networkPolicy.Name), &networkPolicy))
	}
	// Response #3 (and #4)
	// If there is no networkPolicy available (or there are more than 1), just create a new one
	if this.networkPolicyName == resources.RC_EMPTY_NAME && len(this.networkPolicies) != 1 {
		networkPolicy := this.svcKubeFactory.CreateNetworkPolicy(this.serviceName)
		// leave the creation itself to patcher+creator so other CFs can update
		this.svcResourceCache.Set(resources.RC_KEY_NETWORK_POLICY, resources.NewResourceCacheEntry(resources.RC_EMPTY_NAME, networkPolicy))
	}
}

func (this *NetworkPolicyCF) Cleanup() bool {
	// Network Policy should not have any deletion dependencies
	if networkPolicyEntry, networkPolicyExists := this.svcResourceCache.Get(resources.RC_KEY_NETWORK_POLICY); networkPolicyExists {
		if err := this.svcClients.Kube().DeleteNetworkPolicy(networkPolicyEntry.GetValue().(*networking.NetworkPolicy), &meta.DeleteOptions{}); err != nil && !api_errors.IsNotFound(err) {
			this.ctx.GetLog().Error(err, "Could not delete networkPolicy during cleanup.")
			return false
		} else {
			this.svcResourceCache.Remove(resources.RC_KEY_NETWORK_POLICY)
			this.ctx.GetLog().Info("Network Policy has been deleted.")
		}
	}
	return true
}
