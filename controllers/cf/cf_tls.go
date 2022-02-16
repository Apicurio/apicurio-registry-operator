package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strings"
)

var _ loop.ControlFunction = &TLSCF{}

const ENV_JAVA_OPTIONS = "JAVA_OPTIONS"

// =====

const TLS_SECRET_VOLUME_NAME = "registry-tls-cert-and-key"
const TLS_SECRET_VOLUME_PATH = "/etc/" + TLS_SECRET_VOLUME_NAME
const TLS_PORT_NAME = "https"
const TLS_PORT = 8443
const JAVAOPT_QUARKUS_SSL_CERTIFICATE_FILE = "-Dquarkus.http.ssl.certificate.file="
const JAVAOPT_QUARKUS_SSL_CERTIFICATE_KEY_FILE = "-Dquarkus.http.ssl.certificate.key-file="
const JAVAOPT_QUARKUS_INSECURE_REQUESTS = "-Dquarkus.http.insecure-requests="

type TLSCF struct {
	ctx                    *context.LoopContext
	svcResourceCache       resources.ResourceCache
	svcEnvCache            env.EnvCache
	tlsEnabled             bool
	secretName             string
	certFileName           string
	keyFileName            string
	foundVolumeSecretName  string
	foundVolumeSecretItems []core.KeyToPath
	deploymentExists       bool
	deploymentEntry        resources.ResourceCacheEntry
	updateDeployment       bool
	serviceExists          bool
	serviceEntry           resources.ResourceCacheEntry
	servicePorts           []core.ServicePort
	updateService          bool
	javaOptions            []string
	targetJavaOptions      []string
	updateEnv              bool
}

func NewTLSCF(ctx *context.LoopContext) loop.ControlFunction {
	return &TLSCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcEnvCache:      ctx.GetEnvCache(),
		tlsEnabled:       false,
		secretName:       "",
		certFileName:     "tls.crt",
		keyFileName:      "tls.key",
	}
}

func (this *TLSCF) Describe() string {
	return "TLSCF"
}

func (this *TLSCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		tls := specEntry.GetValue().(*ar.ApicurioRegistry).Spec.Configuration.Security.Tls
		this.tlsEnabled = (tls.SecretName != "")
		this.secretName = tls.SecretName
		this.certFileName = "tls.crt"
		this.keyFileName = "tls.key"
		if tls.Certificate != "" {
			this.certFileName = tls.Certificate
		}
		if tls.Key != "" {
			this.keyFileName = tls.Key
		}
	}

	// Observation #2
	// Deployment exists
	this.foundVolumeSecretName = ""

	this.deploymentEntry, this.deploymentExists = this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)
	if this.deploymentExists {
		deployment := this.deploymentEntry.GetValue().(*apps.Deployment)
		for i, v := range deployment.Spec.Template.Spec.Volumes {
			if v.Name == TLS_SECRET_VOLUME_NAME {
				this.foundVolumeSecretName = deployment.Spec.Template.Spec.Volumes[i].VolumeSource.Secret.SecretName
				this.foundVolumeSecretItems = deployment.Spec.Template.Spec.Volumes[i].VolumeSource.Secret.Items
			}
		}
	}

	// Observation #3
	// Service exists
	this.serviceEntry, this.serviceExists = this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
	if this.serviceExists {
		this.servicePorts = this.serviceEntry.GetValue().(*core.Service).Spec.Ports
	}

	// Observation #4
	// JAVA_OPTIONS env var exists and contains the required values
	this.targetJavaOptions = []string{}
	this.javaOptions = []string{}
	if val, exists := this.svcEnvCache.Get(ENV_JAVA_OPTIONS); exists {
		this.javaOptions = strings.Fields(val.GetValue().Value)

		// Remove existing TLS items from targetJavaOptions
		for _, javaOpt := range this.javaOptions {
			if !strings.HasPrefix(javaOpt, JAVAOPT_QUARKUS_SSL_CERTIFICATE_FILE) &&
				!strings.HasPrefix(javaOpt, JAVAOPT_QUARKUS_SSL_CERTIFICATE_KEY_FILE) &&
				!strings.HasPrefix(javaOpt, JAVAOPT_QUARKUS_INSECURE_REQUESTS) {
				this.targetJavaOptions = append(this.targetJavaOptions, javaOpt)
			}
		}

		if this.tlsEnabled {
			// Add the TLS items to the targetJavaOptions slice
			tlsJavaOptions := []string{
				JAVAOPT_QUARKUS_SSL_CERTIFICATE_FILE + TLS_SECRET_VOLUME_PATH + "/" + this.certFileName,
				JAVAOPT_QUARKUS_SSL_CERTIFICATE_KEY_FILE + TLS_SECRET_VOLUME_PATH + "/" + this.keyFileName,
				JAVAOPT_QUARKUS_INSECURE_REQUESTS + "redirect",
			}
			this.targetJavaOptions = append(this.targetJavaOptions, tlsJavaOptions...)
		}
	}
}

func (this *TLSCF) Compare() bool {

	// Condition #1
	// Check the supplied secret name has changed
	// Condition #2
	// Check volume secret items contain the cert and key names
	this.updateDeployment = this.deploymentExists &&
		(this.secretName != this.foundVolumeSecretName || this.ShouldUpdateVolumeSecretItems())

	// Condition #3
	// Check service ports includes the TLS port
	this.updateService = this.serviceExists && this.ShouldUpateServicePorts()

	// Condition #4
	// Check deployment envs contains the correct JAVA_OPTIONS
	this.updateEnv = this.ShouldUpateEnv()

	return this.updateDeployment || this.updateService || this.updateEnv
}

func (this *TLSCF) ShouldUpdateVolumeSecretItems() bool {
	certFileNameFound := false
	keyFileNameFound := false
	for _, item := range this.foundVolumeSecretItems {
		if item.Key == this.certFileName {
			certFileNameFound = true
		}
		if item.Key == this.keyFileName {
			keyFileNameFound = true
		}
	}

	// If tls is enabled and cert or key names do not match, update the deployment
	// If tls is disabled and cert or key names do match, update the deployment
	// Otherwise do not update.
	if this.tlsEnabled {
		return !certFileNameFound || !keyFileNameFound
	}
	return certFileNameFound || keyFileNameFound
}

func (this *TLSCF) ShouldUpateServicePorts() bool {
	httpsServicePortFound := false
	for _, port := range this.servicePorts {
		if port.Name == TLS_PORT_NAME {
			httpsServicePortFound = true
		}
	}

	// If tls is enabled and service port is not found, update the service
	// If tls is disabled and service port is found, update the service
	// Otherwise do not update.
	return this.tlsEnabled != httpsServicePortFound
}

func (this *TLSCF) ShouldUpateEnv() bool {
	return this.javaOptions != nil && this.targetJavaOptions != nil && !reflect.DeepEqual(this.javaOptions, this.targetJavaOptions)
}

func (this *TLSCF) Respond() {
	if this.updateDeployment {
		this.AddSecretVolumePatch(this.deploymentEntry, this.secretName, TLS_SECRET_VOLUME_NAME, this.certFileName, this.keyFileName)
		this.AddSecretMountPatch(this.deploymentEntry, TLS_SECRET_VOLUME_NAME, TLS_SECRET_VOLUME_PATH)
		this.AddContainerPortPatch(this.deploymentEntry)
	}
	if this.updateService {
		this.AddServicePortPatch(this.serviceEntry)
	}
	if this.updateEnv {
		this.UpdateEnv()
	}
}

func (this *TLSCF) UpdateEnv() {
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_JAVA_OPTIONS, strings.Join(this.targetJavaOptions, " ")))
}

func (this *TLSCF) AddSecretVolumePatch(deploymentEntry resources.ResourceCacheEntry, secretName string, volumeName string, certFileName string, keyFileName string) {
	deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		volume := core.Volume{
			Name: volumeName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: secretName,
					Items: []core.KeyToPath{
						core.KeyToPath{
							Key:  certFileName,
							Path: certFileName,
						},
						core.KeyToPath{
							Key:  keyFileName,
							Path: keyFileName,
						},
					},
				},
			},
		}

		newVolumes := []core.Volume{}
		for _, v := range deployment.Spec.Template.Spec.Volumes {
			if v.Name != volumeName {
				newVolumes = append(newVolumes, v)
			}
		}
		if this.tlsEnabled {
			newVolumes = append(newVolumes, volume)
		}
		deployment.Spec.Template.Spec.Volumes = newVolumes
		return deployment
	})
}

func (this *TLSCF) AddSecretMountPatch(deploymentEntry resources.ResourceCacheEntry, volumeName string, mountPath string) {
	deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for ci, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				mount := core.VolumeMount{
					Name:      volumeName,
					ReadOnly:  true,
					MountPath: mountPath,
				}

				newVolumeMounts := []core.VolumeMount{}
				for _, v := range deployment.Spec.Template.Spec.Containers[ci].VolumeMounts {
					if v.Name != volumeName {
						newVolumeMounts = append(newVolumeMounts, v)
					}
				}
				if this.tlsEnabled {
					newVolumeMounts = append(newVolumeMounts, mount)
				}
				deployment.Spec.Template.Spec.Containers[ci].VolumeMounts = newVolumeMounts
			}
		}
		return deployment
	})
}

func (this *TLSCF) AddContainerPortPatch(deploymentEntry resources.ResourceCacheEntry) {
	deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for ci, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				containerPort := core.ContainerPort{
					ContainerPort: TLS_PORT,
					Protocol:      core.ProtocolTCP,
				}

				newContainerPorts := []core.ContainerPort{}
				for _, c := range deployment.Spec.Template.Spec.Containers[ci].Ports {
					if c.ContainerPort != TLS_PORT {
						newContainerPorts = append(newContainerPorts, c)
					}
				}
				if this.tlsEnabled {
					newContainerPorts = append(newContainerPorts, containerPort)
				}
				deployment.Spec.Template.Spec.Containers[ci].Ports = newContainerPorts
			}
		}
		return deployment
	})
}

func (this *TLSCF) AddServicePortPatch(serviceEntry resources.ResourceCacheEntry) {
	serviceEntry.ApplyPatch(func(value interface{}) interface{} {
		service := value.(*core.Service).DeepCopy()
		servicePort := core.ServicePort{
			Name:       TLS_PORT_NAME,
			Port:       TLS_PORT,
			TargetPort: intstr.FromInt(TLS_PORT),
			Protocol:   core.ProtocolTCP,
		}

		newServicePorts := []core.ServicePort{}
		for _, p := range service.Spec.Ports {
			if p.Name != TLS_PORT_NAME {
				newServicePorts = append(newServicePorts, p)
			}
		}
		if this.tlsEnabled {
			newServicePorts = append(newServicePorts, servicePort)
		}
		service.Spec.Ports = newServicePorts
		return service
	})
}

func (this *TLSCF) Cleanup() bool {
	// No cleanup
	return true
}
