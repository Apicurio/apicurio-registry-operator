package apicurioregistry

import (
	ocp_apps "github.com/openshift/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type OCPFactory struct {
	ctx *Context
}

func NewOCPFactory(ctx *Context) *OCPFactory {
	return &OCPFactory{
		ctx: ctx,
	}
}

func (this *OCPFactory) GetLabels() map[string]string {
	return map[string]string{
		"app": this.ctx.GetConfiguration().GetAppName(),
	};
}

func (this *OCPFactory) createObjectMeta(typeTag string) meta.ObjectMeta {
	return meta.ObjectMeta{
		GenerateName: this.ctx.GetConfiguration().GetAppName() + "-" + typeTag + "-",
		Namespace:    this.ctx.GetConfiguration().GetAppNamespace(),
		Labels:       this.GetLabels(),
	}
}

func (this *OCPFactory) CreateDeployment() *ocp_apps.DeploymentConfig {
	var terminationGracePeriodSeconds int64 = 30

	return &ocp_apps.DeploymentConfig{
		ObjectMeta: this.createObjectMeta("deployment"),
		Spec: ocp_apps.DeploymentConfigSpec{
			Replicas: *this.ctx.GetConfiguration().GetConfigInt32P(CFG_DEP_REPLICAS),
			Selector: this.GetLabels(),
			Template: &core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: this.GetLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Name:  this.ctx.GetConfiguration().GetAppName(),
						Image: this.ctx.GetConfiguration().GetImage(),
						Ports: []core.ContainerPort{
							{
								ContainerPort: 8080,
								Protocol:      "TCP",
							},
						},
						Env: this.ctx.GetConfiguration().GetEnv(),
						Resources: core.ResourceRequirements{
							Limits: core.ResourceList{
								core.ResourceCPU:    resource.MustParse(this.ctx.GetConfiguration().GetConfig(CFG_DEP_CPU_LIMIT)),
								core.ResourceMemory: resource.MustParse(this.ctx.GetConfiguration().GetConfig(CFG_DEP_MEMORY_LIMIT)),
							},
							Requests: core.ResourceList{
								core.ResourceCPU:    resource.MustParse(this.ctx.GetConfiguration().GetConfig(CFG_DEP_CPU_REQUESTS)),
								core.ResourceMemory: resource.MustParse(this.ctx.GetConfiguration().GetConfig(CFG_DEP_MEMORY_REQUESTS)),
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
			Strategy: ocp_apps.DeploymentStrategy{
				Type: ocp_apps.DeploymentStrategyTypeRolling,
				RollingParams: &ocp_apps.RollingDeploymentStrategyParams{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
				},
			},
		},
	}
}
