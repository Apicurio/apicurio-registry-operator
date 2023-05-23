package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ loop.ControlFunction = &IngressCF{}

type IngressCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache
	svcClients       *client.Clients
	svcStatus        *status.Status
	svcKubeFactory   *factory.KubeFactory
	isCached         bool
	ingresses        []networking.Ingress
	ingressName      string
	serviceName      string
	disableIngress   bool
}

func NewIngressCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &IngressCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcClients:       ctx.GetClients(),
		svcStatus:        services.GetStatus(),
		svcKubeFactory:   services.GetKubeFactory(),
		isCached:         false,
		ingresses:        make([]networking.Ingress, 0),
		ingressName:      resources.RC_NOT_CREATED_NAME_EMPTY,
		serviceName:      resources.RC_NOT_CREATED_NAME_EMPTY,
		disableIngress:   false,
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *IngressCF) Describe() string {
	return "IngressCF"
}

func (this *IngressCF) Sense() {

	this.disableIngress = false
	// terminate execution if ingress is disabled
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.disableIngress = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.DisableIngress
		// Do cleanup in respond
	}

	// Observation #1
	// Get cached Ingress
	ingressEntry, ingressExists := this.svcResourceCache.Get(resources.RC_KEY_INGRESS)
	if ingressExists {
		this.ingressName = ingressEntry.GetName().Str()
	} else {
		this.ingressName = resources.RC_NOT_CREATED_NAME_EMPTY
	}
	this.isCached = ingressExists

	// Observation #2
	// Get ingress(s) we *should* track
	this.ingresses = make([]networking.Ingress, 0)
	ingresses, err := this.svcClients.Kube().GetIngresses(
		this.ctx.GetAppNamespace(),
		meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetAppName().Str(),
		})
	if err == nil {
		for _, ingress := range ingresses.Items {
			if ingress.GetObjectMeta().GetDeletionTimestamp() == nil {
				this.ingresses = append(this.ingresses, ingress)
			}
		}
	}

	// Observation #3
	// Is there a Service already? It must have been created (has a name)
	if serviceEntry, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE); serviceExists {
		service := serviceEntry.GetValue().(*core.Service).Spec
		foundHttpPort := false
		for _, port := range service.Ports {
			if port.Port == HttpPort {
				foundHttpPort = true
			}
		}
		// Disable ingress if there is no HTTP port in the service
		this.disableIngress = this.disableIngress || !foundHttpPort
		this.serviceName = serviceEntry.GetName().Str()
	} else {
		this.serviceName = resources.RC_NOT_CREATED_NAME_EMPTY
	}

	// Observation #4
	// See if the host in the config spec is not empty
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		this.disableIngress = this.disableIngress || specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host == ""
	}

	if this.disableIngress {
		this.log.Debugw("Ingress is disabled")
	} else {
		this.log.Debugw("Ingress is enabled")
	}

	// Update the status
	this.svcStatus.SetConfig(status.CFG_STA_INGRESS_NAME, this.ingressName)
}

func (this *IngressCF) Compare() bool {

	// Condition #1
	// Ingress cached and at the same time it is disabled (or vice versa)
	return (this.isCached == this.disableIngress) &&
		// Condition #2
		// The service has been created
		this.serviceName != resources.RC_NOT_CREATED_NAME_EMPTY
}

func (this *IngressCF) Respond() {
	// Response #1
	// We already know about an ingress (name), and it is in the list
	if this.ingressName != resources.RC_NOT_CREATED_NAME_EMPTY {
		contains := false
		for _, val := range this.ingresses {
			if val.Name == this.ingressName {
				contains = true
				this.svcResourceCache.Set(resources.RC_KEY_INGRESS, resources.NewResourceCacheEntry(common.Name(val.Name), &val))
				break
			}
		}
		if !contains {
			this.ingressName = resources.RC_NOT_CREATED_NAME_EMPTY
		}
	}
	// Response #2
	// Can follow #1, but there must be a single ingress available
	if this.ingressName == resources.RC_NOT_CREATED_NAME_EMPTY && len(this.ingresses) == 1 {
		ingress := this.ingresses[0]
		this.ingressName = ingress.Name
		this.svcResourceCache.Set(resources.RC_KEY_INGRESS, resources.NewResourceCacheEntry(common.Name(ingress.Name), &ingress))
	}
	// Response #3 (and #4)
	// If there is no ingress available (or there are more than 1),
	// create a new one IF not disabled
	if !this.disableIngress && this.ingressName == resources.RC_NOT_CREATED_NAME_EMPTY && len(this.ingresses) != 1 {
		ingress := this.svcKubeFactory.CreateIngress(this.serviceName)
		// leave the creation itself to patcher+creator so other CFs can update
		this.svcResourceCache.Set(resources.RC_KEY_INGRESS, resources.NewResourceCacheEntry(resources.RC_NOT_CREATED_NAME_EMPTY, ingress))
	}

	// Delete an existing ingress if disabled
	if this.disableIngress {
		this.Cleanup()
	}
}

func (this *IngressCF) Cleanup() bool {
	// Ingress should not have any deletion dependencies
	if ingressEntry, ingressExists := this.svcResourceCache.Get(resources.RC_KEY_INGRESS); ingressExists {
		if err := this.svcClients.Kube().DeleteIngress(ingressEntry.GetValue().(*networking.Ingress)); err != nil && !api_errors.IsNotFound(err) {
			this.log.Errorw("could not delete ingress", "error", err)
			return false
		} else {
			this.svcResourceCache.Remove(resources.RC_KEY_INGRESS)
			this.ctx.GetLog().Info("ingress has been deleted.")
		}
	}
	return true
}
