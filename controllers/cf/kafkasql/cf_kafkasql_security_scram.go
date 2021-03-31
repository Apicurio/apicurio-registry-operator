package kafkasql

import (
	ar "github.com/Apicurio/apicurio-registry-operator/api/v2"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

var _ loop.ControlFunction = &KafkasqlSecurityScramCF{}

// REGISTRY_KAFKASQL_CONSUMER_

const ENV_REGISTRY_KAFKASQL_SCRAM_USER = "REGISTRY_KAFKASQL_SCRAM_USER"
const ENV_REGISTRY_KAFKASQL_SCRAM_PASSWORD = "REGISTRY_KAFKASQL_SCRAM_PASSWORD"

const ENV_REGISTRY_KAFKA_COMMON_SASL_MECHANISM = "REGISTRY_KAFKA_COMMON_SASL_MECHANISM"
const ENV_REGISTRY_KAFKA_COMMON_SASL_JAAS_CONFIG = "REGISTRY_KAFKA_COMMON_SASL_JAAS_CONFIG"

const SCRAM_TRUSTSTORE_SECRET_VOLUME_NAME = "registry-kafkasql-scram-truststore"

type KafkasqlSecurityScramCF struct {
	ctx                          *context.LoopContext
	svcResourceCache             resources.ResourceCache
	svcEnvCache                  env.EnvCache
	persistence                  string
	bootstrapServers             string
	truststoreSecretName         string
	valid                        bool
	foundTruststoreSecretName    string
	deploymentExists             bool
	deploymentEntry              resources.ResourceCacheEntry
	scramUser                    string
	scramPasswordSecretName      string
	scramMechanism               string
	foundScramUser               string
	foundScramPasswordSecretName string
	foundScramMechanism          string
}

func NewKafkasqlSecurityScramCF(ctx *context.LoopContext) loop.ControlFunction {
	return &KafkasqlSecurityScramCF{
		ctx:                          ctx,
		svcResourceCache:             ctx.GetResourceCache(),
		svcEnvCache:                  ctx.GetEnvCache(),
		persistence:                  "",
		bootstrapServers:             "",
		truststoreSecretName:         "",
		valid:                        false,
		foundTruststoreSecretName:    "",
		scramUser:                    "",
		scramPasswordSecretName:      "",
		scramMechanism:               "",
		foundScramUser:               "",
		foundScramPasswordSecretName: "",
		foundScramMechanism:          "",
	}
}

func (this *KafkasqlSecurityScramCF) Describe() string {
	return "KafkasqlSecurityScramCF"
}

func (this *KafkasqlSecurityScramCF) Sense() {
	// Observation #1
	// Read the config values
	if specEntry, exists := this.svcResourceCache.Get(resources.RC_KEY_SPEC); exists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry)
		this.persistence = spec.Spec.Configuration.Persistence
		this.bootstrapServers = spec.Spec.Configuration.Kafkasql.BootstrapServers

		this.truststoreSecretName = spec.Spec.Configuration.Kafkasql.Security.Scram.TruststoreSecretName
		this.scramUser = spec.Spec.Configuration.Kafkasql.Security.Scram.User
		this.scramPasswordSecretName = spec.Spec.Configuration.Kafkasql.Security.Scram.PasswordSecretName
		this.scramMechanism = spec.Spec.Configuration.Kafkasql.Security.Scram.Mechanism
	}

	if this.scramMechanism == "" {
		this.scramMechanism = "SCRAM-SHA-512"
	}

	// Observation #2
	// Deployment exists
	this.foundTruststoreSecretName = ""

	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)
	if deploymentExists {
		deployment := deploymentEntry.GetValue().(*apps.Deployment)
		for i, v := range deployment.Spec.Template.Spec.Volumes {
			if v.Name == SCRAM_TRUSTSTORE_SECRET_VOLUME_NAME {
				this.foundTruststoreSecretName = deployment.Spec.Template.Spec.Volumes[i].VolumeSource.Secret.SecretName
			}
		}
	}
	this.deploymentExists = deploymentExists
	this.deploymentEntry = deploymentEntry

	if entry, exists := this.svcEnvCache.Get(ENV_REGISTRY_KAFKASQL_SCRAM_USER); exists {
		this.foundScramUser = entry.GetValue().Value
	}
	if entry, exists := this.svcEnvCache.Get(ENV_REGISTRY_KAFKASQL_SCRAM_PASSWORD); exists {
		this.foundScramPasswordSecretName = entry.GetValue().ValueFrom.SecretKeyRef.Name
	}

	mech := ""
	if entry, exists := this.svcEnvCache.Get(ENV_REGISTRY_KAFKA_COMMON_SASL_MECHANISM); exists {
		mech = entry.GetValue().Value
	}

	// Observation #3
	// Validate the config values
	this.valid = this.persistence == PERSISTENCE_ID && this.bootstrapServers != "" &&
		this.truststoreSecretName != "" &&
		this.scramUser != "" &&
		this.scramPasswordSecretName != ""

	this.foundScramMechanism = mech
	// We won't actively delete old env values if not used
}

func (this *KafkasqlSecurityScramCF) Compare() bool {
	// Condition #1
	return this.valid && (this.truststoreSecretName != this.foundTruststoreSecretName ||
		this.scramUser != this.foundScramUser ||
		this.scramPasswordSecretName != this.foundScramPasswordSecretName ||
		this.scramMechanism != this.foundScramMechanism)
}

func (this *KafkasqlSecurityScramCF) Respond() {
	this.AddEnv(this.truststoreSecretName, SCRAM_TRUSTSTORE_SECRET_VOLUME_NAME,
		this.scramUser, this.scramPasswordSecretName, this.scramMechanism)

	this.AddSecretVolumePatch(this.deploymentEntry, this.truststoreSecretName, SCRAM_TRUSTSTORE_SECRET_VOLUME_NAME)

	this.AddSecretMountPatch(this.deploymentEntry, SCRAM_TRUSTSTORE_SECRET_VOLUME_NAME, "etc/"+SCRAM_TRUSTSTORE_SECRET_VOLUME_NAME)
}

func (this *KafkasqlSecurityScramCF) AddEnv(truststoreSecretName string, truststoreSecretVolumeName string,
	scramUser string, scramPasswordSecretName string, scramMechanism string) {

	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_PROPERTIES_PREFIX, "REGISTRY_"))

	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KAFKASQL_SCRAM_USER, scramUser))
	this.svcEnvCache.Set(env.NewEnvCacheEntry(&core.EnvVar{
		Name: ENV_REGISTRY_KAFKASQL_SCRAM_PASSWORD,
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: scramPasswordSecretName,
				},
				Key: "password",
			},
		},
	}))

	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KAFKA_COMMON_SASL_MECHANISM, scramMechanism))

	jaasConfig := "org.apache.kafka.common.security.scram.ScramLoginModule required username='$(" + ENV_REGISTRY_KAFKASQL_SCRAM_USER +
		")' password='$(" + ENV_REGISTRY_KAFKASQL_SCRAM_PASSWORD + ")';"

	jaasconfigEntry := env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KAFKA_COMMON_SASL_JAAS_CONFIG, jaasConfig)
	jaasconfigEntry.SetInterpolationDependency(ENV_REGISTRY_KAFKASQL_SCRAM_USER)
	jaasconfigEntry.SetInterpolationDependency(ENV_REGISTRY_KAFKASQL_SCRAM_PASSWORD)
	this.svcEnvCache.Set(jaasconfigEntry)

	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KAFKA_COMMON_SECURITY_PROTOCOL, "SASL_SSL"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KAFKA_COMMON_SSL_TRUSTSTORE_TYPE, "PKCS12"))
	this.svcEnvCache.Set(env.NewSimpleEnvCacheEntry(ENV_REGISTRY_KAFKA_COMMON_SSL_TRUSTSTORE_LOCATION,
		"/etc/"+truststoreSecretVolumeName+"/ca.p12"))
	this.svcEnvCache.Set(env.NewEnvCacheEntry(&core.EnvVar{
		Name: ENV_REGISTRY_KAFKA_COMMON_SSL_TRUSTSTORE_PASSWORD,
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

func (this *KafkasqlSecurityScramCF) AddSecretVolumePatch(deploymentEntry resources.ResourceCacheEntry, secretName string, volumeName string) {
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

func (this *KafkasqlSecurityScramCF) AddSecretMountPatch(deploymentEntry resources.ResourceCacheEntry, volumeName string, mountPath string) {
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

func (this *KafkasqlSecurityScramCF) Cleanup() bool {
	// No cleanup
	return true
}
