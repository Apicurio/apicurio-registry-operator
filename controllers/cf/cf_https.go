package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
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
const HttpPort = 8080

type HttpsCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache
	svcEnvCache      env.EnvCache
	svcClients       *client.Clients
	services         services.LoopServices

	httpsEnabled bool

	secretExists       bool
	targetSecretName   string
	previousSecretName string

	networkPolicyHttpsPortExists bool

	javaOptions       map[string]string
	targetJavaOptions map[string]string
	javaOptionsExists bool

	serviceHttpsPortExists bool

	secretVolumeExists       bool
	secretVolumeMountExists  bool
	containerHttpsPortExists bool
	networkPolicyExists      bool

	httpEnabled                 bool
	serviceHttpPortExists       bool
	containerHttpPortExists     bool
	networkPolicyHttpPortExists bool
}

func NewHttpsCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &HttpsCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		svcClients:       ctx.GetClients(),
		services:         services,
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
	this.httpEnabled = true
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := entry.GetValue().(*ar.ApicurioRegistry).Spec
		this.targetSecretName = spec.Configuration.Security.Https.SecretName
		this.httpEnabled = !spec.Configuration.Security.Https.DisableHttp
	}
	this.log.Debugw("Observation #1", "this.targetSecretName", this.targetSecretName,
		"this.httpEnabled", this.httpEnabled)

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
				this.services.GetConditionManager().GetConfigurationErrorCondition().TransitionInvalid(this.targetSecretName, "spec.configuration.security.https.secretName")
				// No need to transition to not ready, since we can just run without HTTP
				this.ctx.SetRequeueDelaySec(10)
			} else {
				this.secretExists = true
			}
		} else {
			this.log.Errorw("HTTPS secret referenced in Apicurio Registry CR is missing",
				"secretName", this.targetSecretName, "error", err)
			this.services.GetConditionManager().GetConfigurationErrorCondition().TransitionInvalid(this.targetSecretName, "spec.configuration.security.https.secretName")
			// No need to transition to not ready, since we can just run without HTTP
			this.ctx.SetRequeueDelaySec(10)
		}
	}
	this.log.Debugw("Observation #2", "this.targetSecretName", this.targetSecretName,
		"this.secretExists", this.secretExists)

	// Observation #3
	// Get cached service and check if HTTPS port is enabled
	this.serviceHttpsPortExists = false
	this.serviceHttpPortExists = false
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE); exists {
		service := entry.GetValue().(*core.Service).Spec
		for _, port := range service.Ports {
			if port.Port == HttpsPort {
				this.serviceHttpsPortExists = true
			}
			if port.Port == HttpPort {
				this.serviceHttpPortExists = true
			}
		}
	}
	this.log.Debugw("Observation #3", "this.serviceHttpsPortExists", this.serviceHttpsPortExists,
		"this.serviceHttpPortExists", this.serviceHttpPortExists)

	// Observation #4
	// Check deployment has mounted the secret from the config as a volume
	this.secretVolumeExists = false
	this.secretVolumeMountExists = false
	this.containerHttpsPortExists = false
	this.containerHttpPortExists = false
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists {
		deployment := entry.GetValue().(*apps.Deployment)
		container := common.GetContainerByName(deployment.Spec.Template.Spec.Containers, factory.REGISTRY_CONTAINER_NAME)

		if this.targetSecretName != "" {
			// Volume
			for _, volume := range deployment.Spec.Template.Spec.Volumes {
				if volume.Name == this.targetSecretName {
					this.secretVolumeExists = true
					break
				}
			}
			// Volume mount
			for _, mount := range container.VolumeMounts {
				if mount.Name == this.targetSecretName {
					this.secretVolumeMountExists = true
					break
				}
			}
		}
		// Container port
		for _, port := range container.Ports {
			if port.ContainerPort == HttpsPort {
				this.containerHttpsPortExists = true
			}
			if port.ContainerPort == HttpPort {
				this.containerHttpPortExists = true
			}
		}
	}
	this.log.Debugw("Observation #4", "this.secretVolumeExists", this.secretVolumeExists,
		"this.secretVolumeMountExists", this.secretVolumeMountExists,
		"this.containerHttpsPortExists", this.containerHttpsPortExists,
		"this.containerHttpPortExists", this.containerHttpPortExists)

	// Observation #5
	// Find out if Java options is set
	this.targetJavaOptions = map[string]string{
		"-Dquarkus.http.ssl.certificate.file":     "/certs/tls.crt",
		"-Dquarkus.http.ssl.certificate.key-file": "/certs/tls.key",
	}

	var err error = nil
	if this.javaOptions, err = env.ParseJavaOptionsMap(this.svcEnvCache); err == nil {
		this.javaOptionsExists = true
		for k, v := range this.targetJavaOptions {
			vold, exists := this.javaOptions[k]
			this.javaOptionsExists = this.javaOptionsExists && exists && v == vold
		}
	} else {
		this.javaOptions = nil
		this.log.Errorw("could not parse env. variables "+env.JAVA_OPTIONS+" or "+env.JAVA_OPTIONS_LEGACY, "error", err)
	}

	this.log.Debugw("Observation #5", "this.javaOptionsExists", this.javaOptionsExists,
		"this.javaOptions", this.javaOptions)

	// Observation #6
	// Find out if the HTTPS port is set in the NetworkPolicy
	// NOTE: Network policy may not be created immediately, so we need to wait
	// until it is available.
	this.networkPolicyHttpsPortExists = false
	this.networkPolicyHttpPortExists = false
	this.networkPolicyExists = false
	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_NETWORK_POLICY); exists {
		policy := entry.GetValue().(*networking.NetworkPolicy)
		this.networkPolicyExists = true
		for _, rule := range policy.Spec.Ingress {
			for _, port := range rule.Ports {
				if *port.Protocol == "TCP" && int(port.Port.IntValue()) == HttpsPort {
					this.networkPolicyHttpsPortExists = true
				}
				if *port.Protocol == "TCP" && int(port.Port.IntValue()) == HttpPort {
					this.networkPolicyHttpPortExists = true
				}
			}
		}
	}
	this.log.Debugw("Observation #6", "this.networkPolicyHttpsPortExists", this.networkPolicyHttpsPortExists,
		"this.networkPolicyHttpPortExists", this.networkPolicyHttpPortExists)
}

func (this *HttpsCF) Compare() bool {
	this.httpsEnabled = this.targetSecretName != "" && this.secretExists                       // Observation #1, #2
	actionNeeded := (this.secretExists && this.targetSecretName != this.previousSecretName) || // Secret renamed or removed
		(this.httpsEnabled != this.serviceHttpsPortExists) || // Observation #3
		(this.httpsEnabled != this.secretVolumeExists) || // Observation #4
		(this.httpsEnabled != this.secretVolumeMountExists) || // Observation #4
		(this.httpsEnabled != this.containerHttpsPortExists) || // Observation #4
		(this.httpsEnabled != this.javaOptionsExists) || // Observation #5
		(this.networkPolicyExists && this.httpsEnabled != this.networkPolicyHttpsPortExists) || // Observation #6
		// HTTP port
		((!this.httpsEnabled || this.httpEnabled) != this.serviceHttpPortExists) ||
		((!this.httpsEnabled || this.httpEnabled) != this.containerHttpPortExists) ||
		(this.networkPolicyExists && (!this.httpsEnabled || this.httpEnabled) != this.networkPolicyHttpPortExists)
	valid := this.javaOptions != nil
	return valid && actionNeeded
}

func (this *HttpsCF) Respond() {

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT); exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			deployment := value.(*apps.Deployment).DeepCopy()
			apicurioContainer := common.GetContainerByName(deployment.Spec.Template.Spec.Containers, factory.REGISTRY_CONTAINER_NAME)

			httpsPort := &core.ContainerPort{
				ContainerPort: HttpsPort,
				Protocol:      core.ProtocolTCP,
			}

			httpPort := &core.ContainerPort{
				ContainerPort: HttpPort,
				Protocol:      core.ProtocolTCP,
			}

			if this.httpsEnabled && !this.secretVolumeExists {
				// Remove old secret reference if exists
				if this.previousSecretName != "" {
					common.RemoveVolumeFromDeployment(deployment, NewSecretVolume(this.previousSecretName))
				}
				common.SetVolumeInDeployment(deployment, NewSecretVolume(this.targetSecretName))
				this.log.Debugw("added secret volume")
			}
			if !this.httpsEnabled && this.secretVolumeExists {
				common.RemoveVolumeFromDeployment(deployment, NewSecretVolume(this.targetSecretName))
				this.log.Debugw("removed secret volume")
			}
			if this.httpsEnabled && !this.secretVolumeMountExists {
				// Remove old secret reference if exists
				if this.previousSecretName != "" {
					common.RemoveVolumeMountFromContainer(apicurioContainer, NewSecretVolumeMount(this.previousSecretName))
				}
				common.AddVolumeMountToContainer(apicurioContainer, NewSecretVolumeMount(this.targetSecretName))
				this.log.Debugw("added secret volume mount")
			}
			if !this.httpsEnabled && this.secretVolumeMountExists {
				common.RemoveVolumeMountFromContainer(apicurioContainer, NewSecretVolumeMount(this.targetSecretName))
				this.log.Debugw("removed secret volume mount")
			}
			if this.httpsEnabled && !this.containerHttpsPortExists {
				common.AddContainerPort(&apicurioContainer.Ports, httpsPort)
				this.log.Debugw("added container HTTPS port")
			}
			if !this.httpsEnabled && this.containerHttpsPortExists {
				common.RemovePortFromContainer(apicurioContainer, httpsPort)
				this.log.Debugw("removed container HTTPS port")
			}
			// HTTP port
			if (!this.httpsEnabled || this.httpEnabled) && !this.containerHttpPortExists {
				common.AddContainerPort(&apicurioContainer.Ports, httpPort)
				this.log.Debugw("added container HTTP port")
			}
			if this.httpsEnabled && !this.httpEnabled && this.containerHttpPortExists {
				common.RemovePortFromContainer(apicurioContainer, httpPort)
				this.log.Debugw("removed container HTTP port")
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

			httpsPort := &core.ServicePort{
				Name:       "https",
				Protocol:   core.ProtocolTCP,
				Port:       HttpsPort,
				TargetPort: intstr.FromInt(HttpsPort),
			}
			httpPort := &core.ServicePort{
				Name:       "http",
				Protocol:   core.ProtocolTCP,
				Port:       HttpPort,
				TargetPort: intstr.FromInt(HttpPort),
			}

			if this.httpsEnabled && !this.serviceHttpsPortExists {
				common.AddPortToService(service, httpsPort)
				this.log.Debugw("added HTTPS port to service")
			}
			if !this.httpsEnabled && this.serviceHttpsPortExists {
				common.RemovePortFromService(service, httpsPort)
				this.log.Debugw("removed HTTPS port from service")
			}
			// HTTP port
			if (!this.httpsEnabled || this.httpEnabled) && !this.serviceHttpPortExists {
				common.AddPortToService(service, httpPort)
				this.log.Debugw("added HTTP port to service")
			}
			if this.httpsEnabled && !this.httpEnabled && this.serviceHttpPortExists {
				common.RemovePortFromService(service, httpPort)
				this.log.Debugw("removed HTTP port from service")
			}

			return service
		})
	}

	if entry, exists := this.svcResourceCache.Get(resources.RC_KEY_NETWORK_POLICY); exists {
		entry.ApplyPatch(func(value interface{}) interface{} {
			policy := value.(*networking.NetworkPolicy).DeepCopy()

			tcp := core.ProtocolTCP

			httpsRule := &networking.NetworkPolicyIngressRule{
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

			httpRule := &networking.NetworkPolicyIngressRule{
				Ports: []networking.NetworkPolicyPort{
					{
						Protocol: &tcp,
						Port: &intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: HttpPort,
						},
					},
				},
			}

			if this.httpsEnabled && !this.networkPolicyHttpsPortExists {
				common.AddRuleToNetworkPolicy(policy, httpsRule)
				this.log.Debugw("added HTTPS network policy rule")
			}
			if !this.httpsEnabled && this.networkPolicyHttpsPortExists {
				common.RemoveRuleFromNetworkPolicy(policy, httpsRule)
				this.log.Debugw("removed HTTPS network policy rule")
			}
			// HTTP port
			if (!this.httpsEnabled || this.httpEnabled) && !this.networkPolicyHttpPortExists {
				common.AddRuleToNetworkPolicy(policy, httpRule)
				this.log.Debugw("added HTTP network policy rule")
			}
			if this.httpsEnabled && !this.httpEnabled && this.networkPolicyHttpPortExists {
				common.RemoveRuleFromNetworkPolicy(policy, httpRule)
				this.log.Debugw("removed HTTP network policy rule")
			}

			return policy
		})
	}

	if this.secretExists {
		this.previousSecretName = this.targetSecretName
	} else {
		this.previousSecretName = ""
	}
}

func (this *HttpsCF) Cleanup() bool {
	// No cleanup
	return true
}

func NewSecretVolume(name string) *core.Volume {
	return &core.Volume{
		Name: name,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: name,
			},
		},
	}
}

func NewSecretVolumeMount(name string) *core.VolumeMount {
	return &core.VolumeMount{
		Name:      name,
		MountPath: TlsCertMountPath,
		ReadOnly:  true,
	}
}
