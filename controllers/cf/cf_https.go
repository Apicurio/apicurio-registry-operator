package cf

import (
	"errors"
	"fmt"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"time"
)

var _ loop.ControlFunction = &HttpsCF{}

const JavaOptions = "-Dquarkus.http.ssl.certificate.file=/certs/tls.crt -Dquarkus.http.ssl.certificate.key-file=/certs/tls.key"
const TlsCertMountPath = "/certs"
const HttpsPort = 8443
const HttpPort = 8080

type HttpsCF struct {
	ctx              context.LoopContext
	log              *zap.SugaredLogger
	svcResourceCache resources.ResourceCache
	svcClients       *client.Clients
	svcStatus        *status.Status
	svcKubeFactory   *factory.KubeFactory
	serviceExists    bool
	serviceEntry     resources.ResourceCacheEntry
	specExists       bool
	specEntry        resources.ResourceCacheEntry
	deploymentExists bool
	deploymentEntry  resources.ResourceCacheEntry
	secretExists     bool
	secretEntry      resources.ResourceCacheEntry

	httpsEnabled bool
	secret       *core.Secret
	secretName   string
	certificate  string
	key          string

	needsReconcile bool
	cantReconcile  bool
}

func NewHttpsCF(ctx context.LoopContext, services services.LoopServices) loop.ControlFunction {
	res := &HttpsCF{
		ctx:              ctx,
		svcResourceCache: ctx.GetResourceCache(),
		svcClients:       ctx.GetClients(),
		svcStatus:        services.GetStatus(),
		svcKubeFactory:   services.GetKubeFactory(),
		specExists:       false,
		deploymentExists: false,
		secretExists:     false,
		httpsEnabled:     false,
		needsReconcile:   false,
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
	specEntry, specExists := this.svcResourceCache.Get(resources.RC_KEY_SPEC)
	this.specEntry = specEntry
	this.specExists = specExists

	if this.specExists {
		spec := specEntry.GetValue().(*ar.ApicurioRegistry).Spec
		httpsConfig := spec.Configuration.Security.Https

		if this.httpsEnabled != httpsConfig.Enabled {
			this.httpsEnabled = httpsConfig.Enabled
			this.needsReconcile = true
		}

		if this.secretName != httpsConfig.SecretName {
			this.secretName = httpsConfig.SecretName
			this.needsReconcile = true
		}

		if this.certificate != httpsConfig.Certificate {
			this.certificate = httpsConfig.Certificate
			this.needsReconcile = true
		}

		if this.key != httpsConfig.Key {
			this.key = httpsConfig.Key
			this.needsReconcile = true
		}

	}

	// Observation #2
	// Get cached service and check if HTTPS port is enabled
	serviceEntry, serviceExists := this.svcResourceCache.Get(resources.RC_KEY_SERVICE)
	this.serviceEntry = serviceEntry
	this.serviceExists = serviceExists

	if this.httpsEnabled && this.serviceExists {
		service := serviceEntry.GetValue().(*core.Service).Spec
		foundHttpsPort := false
		for _, port := range service.Ports {
			if port.Port == HttpsPort {
				foundHttpsPort = true
			}
		}
		if !foundHttpsPort {
			this.needsReconcile = true
		}
	}

	// Observation #3
	// Get Secret containing the certificate and key
	if this.httpsEnabled {
		secret, err := this.svcClients.Kube().GetSecret(
			this.ctx.GetAppNamespace(), common.Name(this.secretName), &meta.GetOptions{})
		if err == nil {
			this.secret = secret
			// Validate provided secret contains both 'tls.crt' and 'tls.key'
			if !common.SecretHasTLSFields(secret) {
				// Log error and cancel current reconciliation and wait for 10 seconds
				this.log.Errorw("both tls.crt and tls.key must be present", "error", errors.New(fmt.Sprintf("Invalid secret: %s", this.secretName)))
				this.cancelCurrentReconcileAndWait(10)
				return
			}
		} else {
			this.log.Errorw("secret referenced in Apicurio Registry CR is missing.", "error", errors.New(fmt.Sprintf("Secret '%s' is missing, error: %s", this.secretName, err)))
			// Log error and cancel current reconciliation and wait for 10 seconds
			this.cancelCurrentReconcileAndWait(10)
			return
		}
	}

	// Observation #4
	// Get cached deployment
	// If httpsEnabled and deploymentExists, check deployment has mounted the secret from the config as a volume
	deploymentEntry, deploymentExists := this.svcResourceCache.Get(resources.RC_KEY_DEPLOYMENT)
	this.deploymentEntry = deploymentEntry
	this.deploymentExists = deploymentExists

	if this.httpsEnabled && this.deploymentExists && this.secretExists {
		deployment := this.deploymentEntry.GetValue().(*apps.Deployment)
		volumes := deployment.Spec.Template.Spec.Volumes
		foundVolume := false
		for _, volume := range volumes {
			if volume.Name == this.secret.Name {
				foundVolume = true
			}
		}
		if !foundVolume {
			this.needsReconcile = true
		}
	}

}

func (this *HttpsCF) Compare() bool {
	// Condition #1
	// Only reconcile when this.needsReconcile = true
	return this.needsReconcile
}

func (this *HttpsCF) Respond() {

	if this.cantReconcile {
		this.cantReconcile = false
		return
	}

	deployment := this.deploymentEntry.GetValue().(*apps.Deployment).DeepCopy()
	apicurioContainer := &deployment.Spec.Template.Spec.Containers[0]
	apicurioService := this.serviceEntry.GetValue().(*core.Service).DeepCopy()

	if this.httpsEnabled {

		// Add volume containing the certificate and key if it's not already part of the deployment
		common.AddVolumeToDeployment(deployment, &core.Volume{
			Name: this.secretName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: this.secretName,
				},
			},
		})

		// Add volume mount containing the certs to the apicurio container if not already there
		common.AddVolumeMountToContainer(apicurioContainer, &core.VolumeMount{
			Name:      this.secretName,
			MountPath: TlsCertMountPath,
			ReadOnly:  true,
		})

		// Add HTTPS_PORT to the apicurio container if not already there
		common.AddPortToContainer(apicurioContainer, &core.ContainerPort{
			ContainerPort: HttpsPort,
		})

		// Add JAVA_OPTIONS environment variable to the apicurio container if not already there
		common.AddEnvVarToContainer(apicurioContainer, &core.EnvVar{
			Name:  "JAVA_OPTIONS",
			Value: JavaOptions,
		})

		// Add HTTPS_PORT to the apicurio service if not already there
		common.AddPortToService(apicurioService, &core.ServicePort{
			Name:       "https",
			Port:       HttpsPort,
			TargetPort: intstr.IntOrString{Type: 0, IntVal: HttpsPort},
		})

		common.AddPortToService(apicurioService, &core.ServicePort{
			Name:       "http",
			Port:       HttpPort,
			TargetPort: intstr.IntOrString{Type: 0, IntVal: HttpPort},
		})

	} else {

		// Add HTTP_PORT if !httpsEnabled
		common.AddPortToService(apicurioService, &core.ServicePort{
			Name:       "http",
			Port:       HttpPort,
			TargetPort: intstr.IntOrString{Type: 0, IntVal: HttpPort},
		})

		// Remove HTTPS_PORT from apicurio service if !httpsEnabled
		common.RemovePortFromService(apicurioService, &core.ServicePort{
			Name:       "https",
			Port:       HttpsPort,
			TargetPort: intstr.IntOrString{Type: 0, IntVal: HttpsPort},
		})

		// Remove JAVA_OPTIONS environment variable from the apicurio container if present
		common.RemoveEnvVarFromContainer(apicurioContainer, &core.EnvVar{
			Name:  "JAVA_OPTIONS",
			Value: JavaOptions,
		})

	}

	// Patch service resource
	this.serviceEntry.ApplyPatch(func(value interface{}) interface{} {
		return apicurioService
	})

	// Patch deployment resource
	this.deploymentEntry.ApplyPatch(func(value interface{}) interface{} {
		return deployment
	})

	this.needsReconcile = false

}

func (this *HttpsCF) Cleanup() bool {
	// No cleanup
	return true
}

func (this *HttpsCF) cancelCurrentReconcileAndWait(duration time.Duration) {
	time.Sleep(duration * time.Second)
	this.needsReconcile = true
	this.cantReconcile = true
}
