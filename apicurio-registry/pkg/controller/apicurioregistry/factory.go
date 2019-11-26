package apicurioregistry

import (
	ar "github.com/apicurio/apicurio-operators/apicurio-registry/pkg/apis/apicur/v1alpha1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Factory struct {
	ctx *Context
}

// Factory creates new resources for the operator,
// afterwards, they should not be recreated unless there is an otherwise unrecoverable error
func NewFactory(ctx *Context) *Factory {
	return &Factory{
		ctx: ctx,
	}
}

func (this *Factory) GetLabels() map[string]string {
	return map[string]string{
		"app": this.ctx.configuration.GetSpecName(),
	};
}

func (this *Factory) createObjectMeta(typeTag string) meta.ObjectMeta {
	return meta.ObjectMeta{
		GenerateName: this.ctx.configuration.GetSpecName() + "-" + typeTag + "-",
		Namespace:    this.ctx.configuration.GetSpecNamespace(),
		Labels:       this.GetLabels(),
	}
}

func (this *Factory) CreateDeployment() *apps.Deployment {
	var terminationGracePeriodSeconds int64 = 30

	return &apps.Deployment{
		ObjectMeta: this.createObjectMeta("deployment"),
		Spec: apps.DeploymentSpec{
			Replicas: this.ctx.configuration.GetConfigInt32P(CFG_DEP_REPLICAS),
			Selector: &meta.LabelSelector{MatchLabels: this.GetLabels()},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: this.GetLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Name:  this.ctx.configuration.GetSpecName(),
						Image: this.ctx.configuration.GetImage(),
						Ports: []core.ContainerPort{
							{
								ContainerPort: 8080,
								Protocol:      "TCP",
							},
						},
						Env: this.ctx.configuration.getEnv(),
						Resources: core.ResourceRequirements{
							Limits: core.ResourceList{
								core.ResourceCPU:    resource.MustParse(this.ctx.configuration.GetConfig(CFG_DEP_CPU_LIMIT)),
								core.ResourceMemory: resource.MustParse(this.ctx.configuration.GetConfig(CFG_DEP_MEMORY_LIMIT)),
							},
							Requests: core.ResourceList{
								core.ResourceCPU:    resource.MustParse(this.ctx.configuration.GetConfig(CFG_DEP_CPU_REQUESTS)),
								core.ResourceMemory: resource.MustParse(this.ctx.configuration.GetConfig(CFG_DEP_MEMORY_REQUESTS)),
							},
						},
						LivenessProbe: &core.Probe{
							Handler: core.Handler{
								HTTPGet: &core.HTTPGetAction{
									Path: "/health/live",
									Port: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
								},
							},
							InitialDelaySeconds: 5,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
						ReadinessProbe: &core.Probe{
							Handler: core.Handler{
								HTTPGet: &core.HTTPGetAction{
									Path: "/health/ready",
									Port: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
								},
							},
							InitialDelaySeconds: 5,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
						TerminationMessagePath: "/dev/termination-log",
						ImagePullPolicy:        core.PullAlways,
					}},
					RestartPolicy:                 core.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					DNSPolicy:                     core.DNSClusterFirst,
				},
			},
			Strategy: apps.DeploymentStrategy{
				Type: apps.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &apps.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
				},
			},
		},
	}
}

func (this *Factory) CreateService() *core.Service {
	labels := this.GetLabels()
	service := &core.Service{
		ObjectMeta: this.createObjectMeta("service"),
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Protocol:   core.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector:        labels,
			Type:            core.ServiceTypeClusterIP,
			SessionAffinity: core.ServiceAffinityNone,
		},
	}
	return service
}

func (this *Factory) CreateIngress(serviceName string) *v1beta1.Ingress {
	meta := this.createObjectMeta("ingress")
	meta.Annotations = map[string]string{
		"nginx.ingress.kubernetes.io/force-ssl-redirect": "false",
		"nginx.ingress.kubernetes.io/rewrite-target":     "/",
		"nginx.ingress.kubernetes.io/ssl-redirect":       "false",
	}
	res := &v1beta1.Ingress{
		ObjectMeta: meta,
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: this.ctx.configuration.GetConfig(CFG_DEP_ROUTE),
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: serviceName,
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return res
}

func (this *Factory) CreateSpec(spec *ar.ApicurioRegistry) *ar.ApicurioRegistry {
	res := &ar.ApicurioRegistry{
		TypeMeta:   spec.TypeMeta,
		ObjectMeta: spec.ObjectMeta,
		Spec:       spec.Spec,
		Status: ar.ApicurioRegistryStatus{
			Image:          this.ctx.configuration.GetConfig(CFG_STA_IMAGE),
			DeploymentName: this.ctx.configuration.GetConfig(CFG_STA_DEPLOYMENT_NAME),
			ServiceName:    this.ctx.configuration.GetConfig(CFG_STA_SERVICE_NAME),
			IngressName:    this.ctx.configuration.GetConfig(CFG_STA_INGRESS_NAME),
			ReplicaCount:   *this.ctx.configuration.GetConfigInt32P(CFG_STA_REPLICA_COUNT),
			// TODO add the rest
		},
	}
	return res
}
