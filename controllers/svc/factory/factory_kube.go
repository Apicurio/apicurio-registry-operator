package factory

import (
	"os"

	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type KubeFactory struct {
	ctx *context.LoopContext
}

func NewKubeFactory(ctx *context.LoopContext) *KubeFactory {
	return &KubeFactory{
		ctx: ctx,
	}
}

const ENV_REGISTRY_VERSION = "REGISTRY_VERSION"
const ENV_OPERATOR_NAME = "OPERATOR_NAME"

// MUST NOT be used directly as selector labels, because some of them may change.
func (this *KubeFactory) GetLabels() map[string]string {

	registryVersion := os.Getenv(ENV_REGISTRY_VERSION)
	if registryVersion == "" {
		panic("Could not determine registry version. Environment variable '" + ENV_REGISTRY_VERSION + "' is empty.")
	}
	operatorName := os.Getenv(ENV_OPERATOR_NAME)
	if operatorName == "" {
		panic("Could not determine operator name. Environment variable '" + ENV_OPERATOR_NAME + "' is empty.")
	}
	app := this.ctx.GetAppName().Str()

	return map[string]string{
		"app": app,

		"apicur.io/type":    "apicurio-registry",
		"apicur.io/name":    app,
		"apicur.io/version": registryVersion,

		"app.kubernetes.io/name":     "apicurio-registry",
		"app.kubernetes.io/instance": app,
		"app.kubernetes.io/version":  registryVersion,

		"app.kubernetes.io/managed-by": operatorName,
	}
}

// Selector labels MUST be static/constant in the life of the application.
// Labels that can change during operator/SCV upgrade, such as "apicur.io/version" MUST NOT be used.
func (this *KubeFactory) GetSelectorLabels() map[string]string {
	return map[string]string{
		"app": this.ctx.GetAppName().Str(),
	}
}

func (this *KubeFactory) createObjectMeta(typeTag string) meta.ObjectMeta {
	return meta.ObjectMeta{
		Name:      this.ctx.GetAppName().Str() + "-" + typeTag,
		Namespace: this.ctx.GetAppNamespace().Str(),
		Labels:    this.GetLabels(),
	}
}

func (this *KubeFactory) CreateDeployment() *apps.Deployment {
	var terminationGracePeriodSeconds int64 = 30
	var replicas int32 = 1

	return &apps.Deployment{
		ObjectMeta: this.createObjectMeta("deployment"),
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &meta.LabelSelector{MatchLabels: this.GetSelectorLabels()},
			Template: core.PodTemplateSpec{
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
								core.ResourceCPU:    resource.MustParse("500m"),
								core.ResourceMemory: resource.MustParse("512Mi"),
							},
						},
						LivenessProbe: &core.Probe{
							ProbeHandler: core.ProbeHandler{
								HTTPGet: &core.HTTPGetAction{
									Path: "/health/live",
									Port: intstr.FromInt(8080),
								},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
						ReadinessProbe: &core.Probe{
							ProbeHandler: core.ProbeHandler{
								HTTPGet: &core.HTTPGetAction{
									Path: "/health/ready",
									Port: intstr.FromInt(8080),
								},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      5,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
						TerminationMessagePath: "/dev/termination-log",
						ImagePullPolicy:        core.PullAlways,
						VolumeMounts: []core.VolumeMount{
							core.VolumeMount{
								Name:      "tmp",
								MountPath: "/tmp",
							},
						},
					}},
					RestartPolicy:                 core.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					DNSPolicy:                     core.DNSClusterFirst,
					Volumes: []core.Volume{
						core.Volume{
							Name:         "tmp",
							VolumeSource: core.VolumeSource{EmptyDir: &core.EmptyDirVolumeSource{}},
						},
					},
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

func (this *KubeFactory) CreateService() *core.Service {
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
			Selector:        this.GetSelectorLabels(),
			Type:            core.ServiceTypeClusterIP,
			SessionAffinity: core.ServiceAffinityNone,
		},
	}
	return service
}

func (this *KubeFactory) CreateIngress(serviceName string) *networking.Ingress {
	if serviceName == "" {
		panic("Required argument.")
	}
	metaData := this.createObjectMeta("ingress")
	metaData.Annotations = map[string]string{
		"nginx.ingress.kubernetes.io/force-ssl-redirect": "false",
		"nginx.ingress.kubernetes.io/rewrite-target":     "/",
		"nginx.ingress.kubernetes.io/ssl-redirect":       "false",
	}
	pathTypePrefix := networking.PathTypePrefix
	res := &networking.Ingress{
		ObjectMeta: metaData,
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: "",
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: serviceName,
											Port: networking.ServiceBackendPort{
												Number: 8080,
											},
										},
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

func (this *KubeFactory) CreateNetworkPolicy(serviceName string) *networking.NetworkPolicy {
	metaData := this.createObjectMeta("networkpolicy")
	protocolTCP := core.Protocol("TCP")
	res := &networking.NetworkPolicy{
		TypeMeta:   meta.TypeMeta{},
		ObjectMeta: metaData,
		Spec: networking.NetworkPolicySpec{
			PodSelector: meta.LabelSelector{
				MatchLabels: this.GetSelectorLabels()},
			Ingress: []networking.NetworkPolicyIngressRule{
				{
					Ports: []networking.NetworkPolicyPort{
						{
							Protocol: &protocolTCP,
							Port: &intstr.IntOrString{
								Type:   0,
								IntVal: 8080,
								StrVal: "8080",
							},
						},
					},
				},
			},
			Egress:      []networking.NetworkPolicyEgressRule{},
			PolicyTypes: []networking.PolicyType{"Ingress"},
		},
	}
	return res
}

func (this *KubeFactory) CreatePodDisruptionBudget() *policy.PodDisruptionBudget {
	podDisruptionBudget := &policy.PodDisruptionBudget{
		ObjectMeta: this.createObjectMeta("pdb"),
		Spec: policy.PodDisruptionBudgetSpec{
			Selector: &meta.LabelSelector{
				MatchLabels: this.GetSelectorLabels(),
			},
			MaxUnavailable: &intstr.IntOrString{
				IntVal: 1,
			},
		},
	}
	return podDisruptionBudget
}
