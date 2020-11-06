package factory

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	ocp_apps "github.com/openshift/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type OCPFactory struct {
	ctx         *context.LoopContext
	kubeFactory *KubeFactory
}

func NewOCPFactory(ctx *context.LoopContext, kubeFactory *KubeFactory) *OCPFactory {
	return &OCPFactory{
		ctx:         ctx,
		kubeFactory: kubeFactory,
	}
}

func (this *OCPFactory) GetLabels() map[string]string {
	return this.kubeFactory.GetLabels()
}

func (this *OCPFactory) GetSelectorLabels() map[string]string {
	return this.kubeFactory.GetSelectorLabels()
}

func (this *OCPFactory) createObjectMeta(typeTag string) meta.ObjectMeta {
	return this.kubeFactory.createObjectMeta(typeTag)
}

func (this *OCPFactory) CreateDeployment() *ocp_apps.DeploymentConfig {
	var terminationGracePeriodSeconds int64 = 30

	return &ocp_apps.DeploymentConfig{
		ObjectMeta: this.createObjectMeta("deployment"),
		Spec: ocp_apps.DeploymentConfigSpec{
			Replicas: 1,
			Selector: this.GetSelectorLabels(),
			Template: &core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: this.GetLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{{
						Name:  this.ctx.GetAppName().Str(),
						Image: "",
						Ports: []core.ContainerPort{
							{
								ContainerPort: 8080,
								Protocol:      "TCP",
							},
						},
						Env: []core.EnvVar{},
						Resources: core.ResourceRequirements{
							Limits: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("1"),
								core.ResourceMemory: resource.MustParse("1300Mi"),
							},
							Requests: core.ResourceList{
								core.ResourceCPU:    resource.MustParse("0.1"),
								core.ResourceMemory: resource.MustParse("600Mi"),
							},
						},
						LivenessProbe: &core.Probe{
							Handler: core.Handler{
								HTTPGet: &core.HTTPGetAction{
									Path: "/health/live",
									Port: intstr.FromInt(8080),
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
									Port: intstr.FromInt(8080),
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
