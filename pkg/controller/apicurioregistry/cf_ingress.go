package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ ControlFunction = &IngressCF{}

type IngressCF struct {
	ctx               *Context
	isCached          bool
	ingresses         []extensions.Ingress
	ingressName       string
	serviceName       string
	targetHostIsEmpty bool
}

func NewIngressCF(ctx *Context) ControlFunction {

	err := ctx.GetController().Watch(&source.Kind{Type: &extensions.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ar.ApicurioRegistry{},
	})

	if err != nil {
		panic("Error creating Ingress watch.")
	}

	return &IngressCF{
		ctx:               ctx,
		isCached:          false,
		ingresses:         make([]extensions.Ingress, 0),
		ingressName:       RC_EMPTY_NAME,
		serviceName:       RC_EMPTY_NAME,
		targetHostIsEmpty: true,
	}
}

func (this *IngressCF) Describe() string {
	return "IngressCF"
}

func (this *IngressCF) Sense() {

	// Observation #1
	// Get cached Ingress
	ingressEntry, ingressExists := this.ctx.GetResourceCache().Get(RC_KEY_INGRESS)
	if ingressExists {
		this.ingressName = ingressEntry.GetName()
	} else {
		this.ingressName = RC_EMPTY_NAME
	}
	this.isCached = ingressExists

	// Observation #2
	// Get ingress(s) we *should* track
	this.ingresses = make([]extensions.Ingress, 0)
	ingresses, err := this.ctx.GetClients().Kube().GetIngresses(
		this.ctx.GetConfiguration().GetAppNamespace(),
		&meta.ListOptions{
			LabelSelector: "app=" + this.ctx.GetConfiguration().GetAppName(),
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
	serviceEntry, serviceExists := this.ctx.GetResourceCache().Get(RC_KEY_SERVICE)
	if serviceExists {
		this.serviceName = serviceEntry.GetName()
	} else {
		this.serviceName = RC_EMPTY_NAME
	}

	// Observation #4
	// See if the host in the config spec is not empty
	this.targetHostIsEmpty = true
	if specEntry, exists := this.ctx.GetResourceCache().Get(RC_KEY_SPEC); exists {
		this.targetHostIsEmpty = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Host == ""
	}

	// Update the status
	this.ctx.GetConfiguration().SetConfig(CFG_STA_INGRESS_NAME, this.ingressName)
}

func (this *IngressCF) Compare() bool {
	// Condition #1
	// If we already have a ingress cached, skip
	// Condition #2
	// The service has been created
	// Condition #3
	// We will create a new ingress only if the host is not empty
	return !this.isCached &&
		this.serviceName != RC_EMPTY_NAME &&
		!this.targetHostIsEmpty
}

func (this *IngressCF) Respond() {
	// Response #1
	// We already know about a ingress (name), and it is in the list
	if this.ingressName != RC_EMPTY_NAME {
		contains := false
		for _, val := range this.ingresses {
			if val.Name == this.ingressName {
				contains = true
				this.ctx.GetResourceCache().Set(RC_KEY_INGRESS, NewResourceCacheEntry(val.Name, &val))
				break
			}
		}
		if !contains {
			this.ingressName = RC_EMPTY_NAME
		}
	}
	// Response #2
	// Can follow #1, but there must be a single ingress available
	if this.ingressName == RC_EMPTY_NAME && len(this.ingresses) == 1 {
		ingress := this.ingresses[0]
		this.ingressName = ingress.Name
		this.ctx.GetResourceCache().Set(RC_KEY_INGRESS, NewResourceCacheEntry(ingress.Name, &ingress))
	}
	// Response #3 (and #4)
	// If there is no ingress available (or there are more than 1), just create a new one
	if this.ingressName == RC_EMPTY_NAME && len(this.ingresses) != 1 {
		ingress := this.ctx.GetKubeFactory().CreateIngress(this.serviceName)
		// leave the creation itself to patcher+creator so other CFs can update
		this.ctx.GetResourceCache().Set(RC_KEY_INGRESS, NewResourceCacheEntry(RC_EMPTY_NAME, ingress))
	}
}
