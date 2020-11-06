package cf

import (
	"strings"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	ocp_route "github.com/openshift/api/route/v1"
)

var _ loop.ControlFunction = &HostInitRouteOcpCF{}

type HostInitRouteOcpCF struct {
	ctx                             *context.LoopContext
	svcResourceCache                resources.ResourceCache
	isFirstRespond                  bool
	existingHost                    string
	existingRouterCanonicalHostname string
	specEntry                       resources.ResourceCacheEntry
	routeEntry                      resources.ResourceCacheEntry
}

func NewHostInitRouteOcpCF(ctx *context.LoopContext) loop.ControlFunction {
	return &HostInitRouteOcpCF{
		ctx:                             ctx,
		svcResourceCache:                ctx.GetResourceCache(),
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
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.specEntry = specEntry
		this.existingHost = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host
	}

	// Observation #2
	this.existingRouterCanonicalHostname = ""
	if routeEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_ROUTE_OCP); exists {
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

func (this *HostInitRouteOcpCF) Cleanup() bool {
	// No cleanup
	return true
}
