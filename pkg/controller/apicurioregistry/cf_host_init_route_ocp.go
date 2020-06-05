package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	ocp_route "github.com/openshift/api/route/v1"
	"strings"
)

var _ ControlFunction = &HostInitRouteOcpCF{}

type HostInitRouteOcpCF struct {
	ctx                             *Context
	isFirstRespond                  bool
	existingHost                    string
	existingRouterCanonicalHostname string
	specEntry                       ResourceCacheEntry
	routeEntry                      ResourceCacheEntry
}

func NewHostInitRouteOcpCF(ctx *Context) ControlFunction {
	return &HostInitRouteOcpCF{
		ctx:                             ctx,
		isFirstRespond:                  true,
		existingHost:                    "",
		existingRouterCanonicalHostname: "",
		specEntry:                       nil,
		routeEntry:                      nil,
	}
}

func (this *HostInitRouteOcpCF) Describe() string {
	return "HostInitRouteOcpCF"
}

func (this *HostInitRouteOcpCF) Sense() {
	// Optimization
	if !this.isFirstRespond {
		return
	}

	// Observation #1
	// Get spec for patching & the target host
	this.existingHost = ""
	if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
		this.specEntry = specEntry
		this.existingHost = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
	}

	// Observation #2
	this.existingRouterCanonicalHostname = ""
	if routeEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_ROUTE_OCP); exists {
		this.routeEntry = routeEntry
		for _, v := range routeEntry.GetValue().(*ocp_route.Route).Status.Ingress {
			if v.Host == this.existingHost {
				this.existingRouterCanonicalHostname = v.RouterCanonicalHostname
			}
		}
	}
}

func (this *HostInitRouteOcpCF) Compare() bool {
	// Condition #1
	return this.isFirstRespond && this.existingHost != "" && this.existingRouterCanonicalHostname != ""
}

func (this *HostInitRouteOcpCF) Respond() {
	this.isFirstRespond = false
	if !strings.HasSuffix(this.existingHost, this.existingRouterCanonicalHostname) {
		// Response #1
		// Patch the resource
		this.specEntry.ApplyPatch(func(value interface{}) interface{} {
			spec := value.(*ar.ApicurioRegistry).DeepCopy()
			spec.Spec.Deployment.Host = this.existingHost + "." + this.existingRouterCanonicalHostname
			return spec
		})
	}
}
