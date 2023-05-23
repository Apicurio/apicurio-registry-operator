package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ loop.ControlFunction = &HttpsCF{}

const TlsCertMountPath = "/certs"
const HttpsPort = 8443

// TODO
// const HttpPort = 8080

type HttpsCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	svcClients       *client.Clients

	httpsEnabled bool

	secretExists       bool
	targetSecretName   string
	previousSecretName string

	networkPolicyHttpsPortExists bool

	javaOptions       map[string]string
	targetJavaOptions map[string]string
	javaOptionsExists bool

	serviceHttpsPortExists bool

	secretVolumeExists      bool
	secretVolumeMountExists bool
	containerPortExists     bool
	networkPolicyExists     bool
}

func NewHttpsCF(ctx context.LoopContext, _ services.LoopServices) loop.ControlFunction {
	res := &HttpsCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		svcClients:       ctx.GetClients(),
	}
	res.log = ctx.GetLog().Sugar().With("cf", res.Describe())
	return res
}

func (this *HttpsCF) Describe() string {
	return "HttpsCF"
}

func (this *HttpsCF) Sense() {

	// Observation #1
	// Read config values from the Apicurio custom resource
	this.targetSecretName = ""
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := entry.GetValue().(*ar.ApicurioRegistry).Spec
		this.targetSecretName = spec.Configuration.Security.Https.SecretName
	}
	this.log.Debugw("Observation #1", "this.targetSecretName", this.targetSecretName)

	// Observation #2
	// Get Secret containing the certificate and key
	this.secretExists = false
	if this.targetSecretName != "" {
		secret, err := this.svcClients.Kube().
			GetSecret(this.ctx.GetAppNamespace(), common.Name(this.targetSecretName), &meta.GetOptions{})

		if err == nil {
			if !common.SecretHasField(secret, "tls.crt") || !common.SecretHasField(secret, "tls.key") {
				this.log.Errorw("HTTPS secret referenced in Apicurio Registry CR must have both tls.crt and tls.key fields",
					"secretName", this.targetSecretName)
				this.ctx.SetRequeueDelaySec(10)
			} else {
				this.secretExists = true
			}
		} else {
			this.log.Errorw("HTTPS secret referenced in Apicurio Registry CR is missing",
				"secretName", this.targetSecretName, "error", err)
			this.ctx.SetRequeueDelaySec(10)
		}
	}
	this.log.Debugw("Observation #2", "this.targetSecretName", this.targetSecretName,
		"this.secretExists", this.secretExists)

	// Observation #3
	// Get cached service and check if HTTPS port is enabled
	this.serviceHttpsPortExists = false
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE); exists {
		service := entry.GetValue().(*core.Service).Spec
		for _, port := range service.Ports {
			if port.Port == HttpsPort {
				this.serviceHttpsPortExists = true
				break
			}
		}
	}
	this.log.Debugw("Observation #3", "this.serviceHttpsPortExists", this.serviceHttpsPortExists)

	// Observation #4
	// Check deployment has mounted the secret from the config as a volume
	this.secretVolumeExists = false
	this.secretVolumeMountExists = false
	this.containerPortExists = false
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists && this.targetSecretName != "" {
		deployment := entry.GetValue().(*apps.Deployment)
		// Volume
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.Name == this.targetSecretName {
				this.secretVolumeExists = true
				break
			}
		}
		// Volume mount
		for _, mount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
			if mount.Name == this.targetSecretName {
				this.secretVolumeMountExists = true
				break
			}
		}
		// Container port
		for _, port := range deployment.Spec.Template.Spec.Containers[0].Ports {
			if port.ContainerPort == HttpsPort {
				this.containerPortExists = true
				break
			}
		}
	}
	this.log.Debugw("Observation #4", "this.secretVolumeExists", this.secretVolumeExists,
		"this.secretVolumeMountExists", this.secretVolumeMountExists,
		"this.containerPortExists", this.containerPortExists)

	// Observation #5
	// Find out if JAVA_OPTIONS is set
	this.targetJavaOptions = map[string]string{
		"-Dquarkus.http.ssl.certificate.file":     "/certs/tls.crt",
		"-Dquarkus.http.ssl.certificate.key-file": "/certs/tls.key",
	}
	this.javaOptions = env.ParseJavaOptionsMap(this.svcEnvCache)
	this.javaOptionsExists = true
	for k, v := range this.targetJavaOptions {
		vold, exists := this.javaOptions[k]
		this.javaOptionsExists = this.javaOptionsExists && exists && v == vold
	}
	this.log.Debugw("Observation #5", "this.javaOptionsExists", this.javaOptionsExists,
		"this.javaOptions", this.javaOptions)

	// Observation #6
	// Find out if the HTTPS port is set in the NetworkPolicy
	// NOTE: Network policy may not be created immediately, so we need to wait
	// until it is available.
	this.networkPolicyHttpsPortExists = false
	this.networkPolicyExists = false
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_NETWORK_POLICY); exists {
		policy := entry.GetValue().(*networking.NetworkPolicy)
		this.networkPolicyExists = true
		for _, rule := range policy.Spec.Ingress {
			for _, port := range rule.Ports {
				if *port.Protocol == "TCP" && int(port.Port.IntValue()) == HttpsPort {
					this.networkPolicyHttpsPortExists = true
				}
			}
		}
	}
	this.log.Debugw("Observation #6", "this.networkPolicyHttpsPortExists", this.networkPolicyHttpsPortExists)
}

func (this *HttpsCF) Compare() bool {
	this.httpsEnabled = this.targetSecretName != "" && this.secretExists // Observation #1, #2

	return (this.targetSecretName != this.previousSecretName) || // Secret renamed or removed
		(this.httpsEnabled != this.serviceHttpsPortExists) || // Observation #3
		(this.httpsEnabled != this.secretVolumeExists) || // Observation #4
		(this.httpsEnabled != this.secretVolumeMountExists) || // Observation #4
		(this.httpsEnabled != this.containerPortExists) || // Observation #4
		(this.httpsEnabled != this.javaOptionsExists) || // Observation #5
		(this.networkPolicyExists && this.httpsEnabled != this.networkPolicyHttpsPortExists) // Observation #6
}

func (this *HttpsCF) Respond() {

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()

			apicurioContainer := &deployment.Spec.Template.Spec.Containers[0]

			volume := &core.Volume{
				Name: this.targetSecretName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: this.targetSecretName,
					},
				},
			}

			volumeMount := &core.VolumeMount{
				Name:      this.targetSecretName,
				MountPath: TlsCertMountPath,
				ReadOnly:  true,
			}

			port := &core.ContainerPort{
				ContainerPort: HttpsPort,
			}

			if this.httpsEnabled && !this.secretVolumeExists {
				common.SetVolumeInDeployment(deployment, volume)
				this.log.Debugw("added secret volume")
			}
			if !this.httpsEnabled && this.secretVolumeExists {
				common.RemoveVolumeFromDeployment(deployment, volume)
				this.log.Debugw("removed secret volume")
			}
			if this.httpsEnabled && !this.secretVolumeMountExists {
				common.AddVolumeMountToContainer(apicurioContainer, volumeMount)
				this.log.Debugw("added secret volume mount")
			}
			if !this.httpsEnabled && this.secretVolumeMountExists {
				common.RemoveVolumeMountFromContainer(apicurioContainer, volumeMount)
				this.log.Debugw("removed secret volume mount")
			}
			if this.httpsEnabled && !this.containerPortExists {
				common.AddPortToContainer(apicurioContainer, port)
				this.log.Debugw("added container port")
			}
			if !this.httpsEnabled && this.containerPortExists {
				common.RemovePortFromContainer(apicurioContainer, port)
				this.log.Debugw("removed container port")
			}

			return deployment
		})
	}

	// Java Options
	if this.httpsEnabled && !this.javaOptionsExists {
		for k, v := range this.targetJavaOptions {
			this.javaOptions[k] = v
		}
		env.SaveJavaOptionsMap(this.svcEnvCache, this.javaOptions)
		this.log.Debugw("added java options", "this.javaOptions", this.javaOptions)
	}
	if !this.httpsEnabled && this.javaOptionsExists {
		changed := false
		for k, _ := range this.targetJavaOptions {
			if _, ok := this.javaOptions[k]; ok {
				delete(this.javaOptions, k)
				changed = true
			}
		}
		if changed {
			env.SaveJavaOptionsMap(this.svcEnvCache, this.javaOptions)
			this.log.Debugw("removed java options", "this.javaOptions", this.javaOptions)
		}
	}

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE); exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			service := value.(*core.Service).DeepCopy()

			port := &core.ServicePort{
				Name:       "https",
				Protocol:   core.ProtocolTCP,
				Port:       HttpsPort,
				TargetPort: intstr.FromInt(HttpsPort),
			}

			if this.httpsEnabled && !this.serviceHttpsPortExists {
				common.AddPortToService(service, port)
				this.log.Debugw("added HTTPS port to service")
			}
			if !this.httpsEnabled && this.serviceHttpsPortExists {
				common.RemovePortFromService(service, port)
				this.log.Debugw("removed HTTPS port from service")
			}

			return service
		})
	}

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_NETWORK_POLICY); exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			policy := value.(*networking.NetworkPolicy).DeepCopy()

			tcp := core.ProtocolTCP

			rule := &networking.NetworkPolicyIngressRule{
				Ports: []networking.NetworkPolicyPort{
					{
						Protocol: &tcp,
						Port: &intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: HttpsPort,
						},
					},
				},
			}

			if this.httpsEnabled && !this.networkPolicyHttpsPortExists {
				common.AddRuleToNetworkPolicy(policy, rule)
				this.log.Debugw("added network policy rule")
			}
			if !this.httpsEnabled && this.networkPolicyHttpsPortExists {
				common.RemoveRuleFromNetworkPolicy(policy, rule)
				this.log.Debugw("removed network policy rule")
			}

			return policy
		})
	}

	this.previousSecretName = this.targetSecretName
}

func (this *HttpsCF) Cleanup() bool {
	// No cleanup
	return true
}
