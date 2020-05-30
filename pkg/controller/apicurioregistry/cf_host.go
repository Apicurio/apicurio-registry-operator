package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	extensions "k8s.io/api/extensions/v1beta1"
)

var _ ControlFunction = &HostCF{}

type HostCF struct {
	ctx           *Context
	ingressEntry  ResourceCacheEntry
	ingressExists bool
	serviceName   string
	existingHost  string
	targetHost    string
}

// This CF makes sure number of host is aligned
// If there is some other way of determining the number of host needed outside of CR,
// modify the Sense stage so this CF knows about it
func NewHostCF(ctx *Context) ControlFunction {
	return &HostCF{
		ctx:           ctx,
		ingressEntry:  nil,
		ingressExists: false,
		serviceName:   RC_EMPTY_NAME,
		existingHost:  RC_EMPTY_NAME,
		targetHost:    RC_EMPTY_NAME,
	}
}

func (this *HostCF) Describe() string {
	return "HostCF"
}

func (this *HostCF) Sense() {

	// Observation #1
	// Get the cached Ingress (if it exists and/or the value)
	ingressEntry, ingressExists := this.ctx.GetResourceCache().Get(RC_KEY_INGRESS)
	this.ingressEntry = ingressEntry
	this.ingressExists = ingressExists

	// Observation #2
	// Is there a Service already? It must have been created (has a name)
	serviceEntry, serviceExists := this.ctx.GetResourceCache().Get(RC_KEY_SERVICE)
	if serviceExists {
		this.serviceName = serviceEntry.GetName() // TODO this may still end up empty, refactor?
	} else {
		this.serviceName = RC_EMPTY_NAME
	}

	// Observation #3
	// Get the existing host (if present)
	this.existingHost = RC_EMPTY_NAME
	if this.ingressExists && this.serviceName != RC_EMPTY_NAME {
		this.existingHost = readHost(this.serviceName, this.ingressEntry.GetValue().(*extensions.Ingress))
	}

	// Observation #4
	// Get target host
	if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
		this.targetHost = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
	}

	// Update state
	this.ctx.GetConfiguration().SetConfig(CFG_STA_ROUTE, this.existingHost)
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
		this.serviceName != RC_EMPTY_NAME &&
		this.existingHost != this.targetHost &&
		this.targetHost != ""
}

func (this *HostCF) Respond() {
	// Response #1
	// Patch the resource
	this.ingressEntry.ApplyPatch(func(value interface{}) interface{} {
		ingress := value.(*extensions.Ingress).DeepCopy()
		writeHost(this.serviceName, ingress, this.targetHost)
		return ingress
	})
}

func readHost(serviceName string, ingress *extensions.Ingress) string {
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.ServiceName == serviceName {
				return rule.Host
			}
		}
	}
	return RC_EMPTY_NAME
}

func writeHost(serviceName string, ingress *extensions.Ingress, host string) {
	for i, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.ServiceName == serviceName {
				ingress.Spec.Rules[i] = extensions.IngressRule{
					Host:             host,
					IngressRuleValue: rule.IngressRuleValue,
				}
			}
		}
	}
}
