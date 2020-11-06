package cf

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ loop.ControlFunction = &StreamsSecurityTLSCF{}

const ENV_REGISTRY_PROPERTIES_PREFIX = "REGISTRY_PROPERTIES_PREFIX"

const ENV_REGISTRY_STREAMS_TOPOLOGY_SECURITY_PROTOCOL = "REGISTRY_STREAMS_TOPOLOGY_SECURITY_PROTOCOL"

const ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_TYPE = "REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_TYPE"
const ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_LOCATION = "REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_LOCATION"
const ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_PASSWORD = "REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_PASSWORD"

const ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_TYPE = "REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_TYPE"
const ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_LOCATION = "REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_LOCATION"
const ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_PASSWORD = "REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_PASSWORD"

const ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SECURITY_PROTOCOL = "REGISTRY_STREAMS_STORAGE-PRODUCER_SECURITY_PROTOCOL"

const ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_KEYSTORE_TYPE = "REGISTRY_STREAMS_STORAGE-PRODUCER_SSL_KEYSTORE_TYPE"
const ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_KEYSTORE_LOCATION = "REGISTRY_STREAMS_STORAGE-PRODUCER_SSL_KEYSTORE_LOCATION"
const ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_KEYSTORE_PASSWORD = "REGISTRY_STREAMS_STORAGE-PRODUCER_SSL_KEYSTORE_PASSWORD"

const ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_TRUSTSTORE_TYPE = "REGISTRY_STREAMS_STORAGE-PRODUCER_SSL_TRUSTSTORE_TYPE"
const ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_TRUSTSTORE_LOCATION = "REGISTRY_STREAMS_STORAGE-PRODUCER_SSL_TRUSTSTORE_LOCATION"
const ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_TRUSTSTORE_PASSWORD = "REGISTRY_STREAMS_STORAGE-PRODUCER_SSL_TRUSTSTORE_PASSWORD"

const KEYSTORE_SECRET_VOLUME_NAME = "registry-streams-tls-keystore"
const TRUSTSTORE_SECRET_VOLUME_NAME = "registry-streams-tls-truststore"

type StreamsSecurityTLSCF struct {
	ctx                       *context.LoopContext
	svcResourceCache          resources.ResourceCache
	svcEnvCache               env.EnvCache
	persistence               string
	bootstrapServers          string
	keystoreSecretName        string
	truststoreSecretName      string
	valid                     bool
	foundKeystoreSecretName   string
	foundTruststoreSecretName string
	deploymentExists          bool
	deploymentEntry           resources.ResourceCacheEntry
}

func NewStreamsSecurityTLSCF(ctx *context.LoopContext) loop.ControlFunction {
	return &StreamsSecurityTLSCF{
		ctx:                       ctx,
		svcResourceCache:          ctx.GetResourceCache(),
		svcEnvCache:               ctx.GetEnvCache(),
		persistence:               "",
		bootstrapServers:          "",
		keystoreSecretName:        "",
		truststoreSecretName:      "",
		valid:                     true,
		foundKeystoreSecretName:   "",
		foundTruststoreSecretName: "",
	}
}

func (this *StreamsSecurityTLSCF) Describe() string {
	return "StreamsSecurityTLSCF"
}

func (this *StreamsSecurityTLSCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		this.persistence = spec.Spec.Configuration.Persistence
		this.bootstrapServers = spec.Spec.Configuration.Streams.BootstrapServers

		this.keystoreSecretName = spec.Spec.Configuration.Streams.Security.Tls.KeystoreSecretName
		this.truststoreSecretName = spec.Spec.Configuration.Streams.Security.Tls.TruststoreSecretName
	}

	// Observation #2
	// Deployment exists
	this.foundKeystoreSecretName = ""
	this.foundTruststoreSecretName = ""

	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)
	if deploymentExists {
		deployment := deploymentEntry.GetValue().(*apps.Deployment)
		for i, v := range deployment.Spec.Template.Spec.Volumes {
			if v.Name == KEYSTORE_SECRET_VOLUME_NAME {
				this.foundKeystoreSecretName = deployment.Spec.Template.Spec.Volumes[i].VolumeSource.Secret.SecretName
			}
			if v.Name == TRUSTSTORE_SECRET_VOLUME_NAME {
				this.foundTruststoreSecretName = deployment.Spec.Template.Spec.Volumes[i].VolumeSource.Secret.SecretName
			}
		}
	}
	this.deploymentExists = deploymentExists
	this.deploymentEntry = deploymentEntry

	// Observation #3
	// Validate the config values
	this.valid = this.persistence == "streams" && this.bootstrapServers != "" &&
		this.keystoreSecretName != "" && this.truststoreSecretName != ""

	// We won't actively delete old env values if not used
}

func (this *StreamsSecurityTLSCF) Compare() bool {
	// Condition #1
	return this.valid && (this.keystoreSecretName != this.foundKeystoreSecretName ||
		this.truststoreSecretName != this.foundTruststoreSecretName)
}

func (this *StreamsSecurityTLSCF) Respond() {
	this.AddEnv(this.keystoreSecretName, KEYSTORE_SECRET_VOLUME_NAME,
		this.truststoreSecretName, TRUSTSTORE_SECRET_VOLUME_NAME)

	this.AddSecretVolumePatch(this.deploymentEntry, this.keystoreSecretName, KEYSTORE_SECRET_VOLUME_NAME)
	this.AddSecretVolumePatch(this.deploymentEntry, this.truststoreSecretName, TRUSTSTORE_SECRET_VOLUME_NAME)

	this.AddSecretMountPatch(this.deploymentEntry, KEYSTORE_SECRET_VOLUME_NAME, "etc/"+KEYSTORE_SECRET_VOLUME_NAME)
	this.AddSecretMountPatch(this.deploymentEntry, TRUSTSTORE_SECRET_VOLUME_NAME, "etc/"+TRUSTSTORE_SECRET_VOLUME_NAME)
}

func (this *StreamsSecurityTLSCF) AddEnv(keystoreSecretName string, keystoreSecretVolumeName string,
	truststoreSecretName string, truststoreSecretVolumeName string) {

	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_PROPERTIES_PREFIX, "REGISTRY_"))

	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_TOPOLOGY_SECURITY_PROTOCOL, "SSL"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_TYPE, "PKCS12"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_LOCATION,
		"/etc/"+keystoreSecretVolumeName+"/user.p12"))
	this.svcEnvCache.Set(env.NewEnvCacheEntry(&core.EnvVar{
		Name: ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_KEYSTORE_PASSWORD,
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: keystoreSecretName,
				},
				Key: "user.password",
			},
		},
	}))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_TYPE, "PKCS12"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_LOCATION,
		"/etc/"+truststoreSecretVolumeName+"/ca.p12"))
	this.svcEnvCache.Set(env.NewEnvCacheEntry(&core.EnvVar{
		Name: ENV_REGISTRY_STREAMS_TOPOLOGY_SSL_TRUSTSTORE_PASSWORD,
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: truststoreSecretName,
				},
				Key: "ca.password",
			},
		},
	}))

	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SECURITY_PROTOCOL, "SSL"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_KEYSTORE_TYPE, "PKCS12"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_KEYSTORE_LOCATION,
		"/etc/"+keystoreSecretVolumeName+"/user.p12"))
	this.svcEnvCache.Set(env.NewEnvCacheEntry(&core.EnvVar{
		Name: ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_KEYSTORE_PASSWORD,
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: keystoreSecretName,
				},
				Key: "user.password",
			},
		},
	}))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_TRUSTSTORE_TYPE, "PKCS12"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_TRUSTSTORE_LOCATION,
		"/etc/"+truststoreSecretVolumeName+"/ca.p12"))
	this.svcEnvCache.Set(env.NewEnvCacheEntry(&core.EnvVar{
		Name: ENV_REGISTRY_STREAMS_STORAGE_PRODUCER_SSL_TRUSTSTORE_PASSWORD,
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: truststoreSecretName,
				},
				Key: "ca.password",
			},
		},
	}))
}

func (this *StreamsSecurityTLSCF) AddSecretVolumePatch(deploymentEntry resources.ResourceCacheEntry, secretName string, volumeName string) {
	deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		volume := core.Volume{
			Name: volumeName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		}
		j := -1
		for i, v := range deployment.Spec.Template.Spec.Volumes {
			if v.Name == volumeName {
				j = i
				deployment.Spec.Template.Spec.Volumes[i] = volume
			}
		}
		if j == -1 {
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
		}
		return deployment
	})
}

func (this *StreamsSecurityTLSCF) AddSecretMountPatch(deploymentEntry resources.ResourceCacheEntry, volumeName string, mountPath string) {
	deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		deployment := value.(*apps.Deployment).DeepCopy()
		for ci, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == this.ctx.GetAppName().Str() {
				mount := core.VolumeMount{
					Name:      volumeName,
					ReadOnly:  true,
					MountPath: mountPath,
				}
				j := -1
				for i, v := range deployment.Spec.Template.Spec.Containers[ci].VolumeMounts {
					if v.Name == volumeName {
						j = i
						deployment.Spec.Template.Spec.Containers[ci].VolumeMounts[i] = mount
					}
				}
				if j == -1 {
					deployment.Spec.Template.Spec.Containers[ci].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[ci].VolumeMounts, mount)
				}
			}
		}
		return deployment
	})
}

func (this *StreamsSecurityTLSCF) Cleanup() bool {
	// No cleanup
	return true
}
