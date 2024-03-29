package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	networking "k8s.io/api/networking/v1"
)

var _ loop.ControlFunction = &HostCF{}

type HostCF struct {
	ctx              context.LoopContext
	svcResourceCache resources.ResourceCache
	svcStatus        *status.Status
	ingressEntry     resources.ResourceCacheEntry
	ingressExists    bool
	serviceName      string
	existingHost     string
	targetHost       string
}

// This CF makes sure number of host is aligned
// If there is some other way of determining the number of host needed outside of CR,
// modify the Sense stage so this CF knows about it
func NewHostCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	return &HostCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcStatus:        services.GetStatus(),
		ingressEntry:     nil,
		ingressExists:    false,
		serviceName:      resources.RC_NOT_CREATED_NAME_EMPTY,
		existingHost:     "",
		targetHost:       "",
	}
}

func (this *HostCF) Describe() string {
	return "HostCF"
}

func (this *HostCF) Sense() {

	// Observation #1
	// Get the cached Ingress (if it exists and/or the value)
	ingressEntry, ingressExists := this.svcResourceCache.Get(resources.RC_KEY_INGRESS)
	this.ingressEntry = ingressEntry
	this.ingressExists = ingressExists

	// Observation #2
	// Is there a Service already? It must have been created (has a name)
	serviceEntry, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
	if serviceExists {
		this.serviceName = serviceEntry.GetName().Str() // TODO this may still end up empty, refactor?
	} else {
		this.serviceName = resources.RC_NOT_CREATED_NAME_EMPTY
	}

	// Observation #3
	// Get the existing host (if present)
	this.existingHost = resources.RC_NOT_CREATED_NAME_EMPTY
	if this.ingressExists && this.serviceName != resources.RC_NOT_CREATED_NAME_EMPTY {
		this.existingHost = readHost(this.serviceName, this.ingressEntry.GetValue().(*networking.Ingress))
	}

	// Observation #4
	// Get target host
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.targetHost = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
	}

	// Update state
	this.svcStatus.SetConfig(status.CFG_STA_ROUTE, this.existingHost)
}

func (this *HostCF) Compare() bool {
	// Condition #1
	// Ingress exists
	// Condition #2
	// Service exists & is created
	// Condition #3
	// Existing host is not the same as the target host (assuming it is never empty)
	// Condition #4
	// Host must not be empty
	return this.ingressEntry != nil &&
		this.serviceName != resources.RC_NOT_CREATED_NAME_EMPTY &&
		this.existingHost != this.targetHost &&
		this.targetHost != ""
}

func (this *HostCF) Respond() {
	// Response #1
	// Patch the resource
	this.ingressEntry.ApplyPatch(func(value interface{}) interface{} {
		ingress := value.(*networking.Ingress).DeepCopy()
		writeHost(this.serviceName, ingress, this.targetHost)
		return ingress
	})
}

func (this *HostCF) Cleanup() bool {
	// No cleanup
	return true
}

func readHost(serviceName string, ingress *networking.Ingress) string {
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service.Name == serviceName {
				return rule.Host
			}
		}
	}
	return resources.RC_NOT_CREATED_NAME_EMPTY
}

func writeHost(serviceName string, ingress *networking.Ingress, host string) {
	for i, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service.Name == serviceName {
				ingress.Spec.Rules[i] = networking.IngressRule{
					Host:             host,
					IngressRuleValue: rule.IngressRuleValue,
				}
			}
		}
	}
}
