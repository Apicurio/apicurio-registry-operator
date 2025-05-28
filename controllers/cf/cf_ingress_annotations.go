package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	networking "k8s.io/api/networking/v1"
)

var _ loop.ControlFunction = &IngressAnnotationsCF{}

type IngressAnnotationsCF struct {
	ctx                       context.LoopContext
	svcResourceCache          resources.ResourceCache
	ingressEntry              resources.ResourceCacheEntry
	ingressEntryExists        bool
	existingAnnotations       map[string]string
	targetAnnotations         map[string]string
	previousTargetAnnotations map[string]string
}

func NewIngressAnnotationsCF(ctx context.LoopContext) loop.ControlFunction {
	return &IngressAnnotationsCF{
		ctx:                       ctx,
		svcResourceCache:          ctx.GetResourceCache(),
		ingressEntry:              nil,
		ingressEntryExists:        false,
		existingAnnotations:       nil,
		targetAnnotations:         nil,
		previousTargetAnnotations: nil,
	}
}

func (this *IngressAnnotationsCF) Describe() string {
	return "IngressAnnotationsCF"
}

func (this *IngressAnnotationsCF) Sense() {
	if this.ingressEntry, this.ingressEntryExists = this.svcResourceCache.Get(resources.RC_KEY_INGRESS); this.ingressEntryExists {
		this.existingAnnotations = this.ingressEntry.GetValue().(*networking.Ingress).Annotations
		if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
			this.targetAnnotations = specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Deployment.Ingress.Annotations
		}
	}
}

func (this *IngressAnnotationsCF) Compare() bool {
	tmp := c.Copy(this.previousTargetAnnotations)
	c.LabelsDelete(&tmp, this.targetAnnotations)
	return this.ingressEntryExists &&
		(!c.LabelsEqual(this.existingAnnotations, this.targetAnnotations) || len(tmp) != 0)
}

func (this *IngressAnnotationsCF) Respond() {
	c.LabelsDelete(&this.previousTargetAnnotations, this.targetAnnotations)
	this.ingressEntry.ApplyPatch(func(value interface{}) interface{} {
		ingress := value.(*networking.Ingress).DeepCopy()
		c.LabelsDeleteUpdate(&ingress.Annotations, this.previousTargetAnnotations, this.targetAnnotations)
		return ingress
	})
	this.previousTargetAnnotations = c.Copy(this.targetAnnotations)
}

func (this *IngressAnnotationsCF) Cleanup() bool {
	// No cleanup
	return true
}
