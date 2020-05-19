package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	extensions "k8s.io/api/extensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ ControlFunction = &HostConfigCF{}

type HostConfigCF struct {
	ctx *Context
}

func NewHostConfigCF(ctx *Context) ControlFunction {
	return &HostConfigCF{ctx: ctx}
}

func (this *HostConfigCF) Describe() string {
	return "Host Configuration"
}

func (this *HostConfigCF) Sense(spec *ar.ApicurioRegistry, request reconcile.Request) error {
	ingress, err := this.ctx.GetClients().Kube().GetCurrentIngress()
	if err == nil {
		_, _, host := extractHost(this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME), ingress)
		if host != nil {
			this.ctx.GetConfiguration().SetConfig(CFG_STA_ROUTE, *host)
		}
	} else {
		this.ctx.GetLog().Error(err, "Warning: Error getting Ingress.")
	}
	return nil
}

func (this *HostConfigCF) Compare(spec *ar.ApicurioRegistry) (bool, error) {
	return this.ctx.GetConfiguration().GetConfig(CFG_STA_ROUTE) != this.ctx.GetConfiguration().GetConfig(CFG_DEP_ROUTE), nil
}

func (this *HostConfigCF) Respond(spec *ar.ApicurioRegistry) (bool, error) {
	this.ctx.GetPatchers().Kube().AddIngressPatch(func(ingress *extensions.Ingress) {
		for i, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.ServiceName == this.ctx.GetConfiguration().GetConfig(CFG_STA_SERVICE_NAME) {
					ingress.Spec.Rules[i] = extensions.IngressRule{
						Host:             this.ctx.GetConfiguration().GetConfig(CFG_DEP_ROUTE),
						IngressRuleValue: rule.IngressRuleValue,
					}
					return
				}
			}
		}
	})
	return true, nil
}

func extractHost(serviceName string, ingress *extensions.Ingress) (*extensions.IngressRule, *extensions.HTTPIngressPath, *string) {
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.ServiceName == serviceName {
				return &rule, &path, &rule.Host
			}
		}
	}
	return nil, nil, nil
}
