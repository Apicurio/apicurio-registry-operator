package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	networking "k8s.io/api/networking/v1"
)

var _ loop.ControlFunction = &IngressClassNameCF{}

type IngressClassNameCF struct {
	ctx                context.LoopContext
	svcResourceCache   resources.ResourceCache
	ingressEntry       resources.ResourceCacheEntry
	ingressEntryExists bool
	existing           string
	target             string
}

func NewIngressClassNameCF(ctx context.LoopContext) loop.ControlFunction {
	return &IngressClassNameCF{
		ctx:                ctx,
		svcResourceCache:   ctx.GetResourceCache(),
		ingressEntry:       nil,
		ingressEntryExists: false,
		existing:           "",
		target:             "",
	}
}

func (this *IngressClassNameCF) Describe() string {
	return "IngressClassNameCF"
}

func (this *IngressClassNameCF) Sense() {
	if this.ingressEntry, this.ingressEntryExists = this.svcResourceCache.Get(resources.RC_KEY_INGRESS); this.ingressEntryExists {
		if existing := this.ingressEntry.GetValue().(*networking.Ingress).Spec.IngressClassName; existing != nil {
			this.existing = *existing
		} else {
			this.existing = ""
		}
		if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
			this.target = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Ingress.IngressClassName
		} else {
			this.target = ""
		}
	}
}

func (this *IngressClassNameCF) Compare() bool {
	return this.ingressEntryExists &&
		this.existing != this.target
}

func (this *IngressClassNameCF) Respond() {
	this.ingressEntry.ApplyPatch(func(value interface{}) interface{} {
		ingress := value.(*networking.Ingress).DeepCopy()
		if this.target != "" {
			ingress.Spec.IngressClassName = &this.target
		} else {
			ingress.Spec.IngressClassName = nil
		}
		return ingress
	})
}

func (this *IngressClassNameCF) Cleanup() bool {
	// No cleanup
	return true
}
