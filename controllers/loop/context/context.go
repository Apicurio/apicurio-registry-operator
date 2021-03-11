package context

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/status"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	sigs_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// A long-lived singleton container for shared, data only, 0 dependencies, components
type LoopContext struct {
	appName      common.Name
	appNamespace common.Namespace
	log          logr.Logger
	requeue      bool

	client sigs_client.Client
	scheme *runtime.Scheme

	resourceCache resources.ResourceCache
	status        *status.Status
	envCache      env.EnvCache
}

// Create a new context when the operator is deployed, provide mostly static data
func NewLoopContext(appName common.Name, appNamespace common.Namespace, log logr.Logger, scheme *runtime.Scheme, client sigs_client.Client) *LoopContext {
	this := &LoopContext{
		appName:      appName,
		appNamespace: appNamespace,
		requeue:      false,
	}
	this.log = log.WithValues("app", appName.Str(), "namespace", appNamespace.Str())

	this.client = client
	this.scheme = scheme

	this.resourceCache = resources.NewResourceCache()
	this.status = status.NewStatus(this.log)
	this.envCache = env.NewEnvCache()

	return this
}

func (this *LoopContext) GetLog() logr.Logger {
	return this.log
}

func (this *LoopContext) GetAppName() common.Name {
	return this.appName
}

func (this *LoopContext) GetAppNamespace() common.Namespace {
	return this.appNamespace
}

func (this *LoopContext) SetRequeue() {
	this.requeue = true
}

func (this *LoopContext) GetAndResetRequeue() bool {
	res := this.requeue
	this.requeue = false
	return res
}

func (this *LoopContext) GetResourceCache() resources.ResourceCache {
	return this.resourceCache
}

func (this *LoopContext) GetClient() sigs_client.Client {
	return this.client
}

func (this *LoopContext) GetScheme() *runtime.Scheme {
	return this.scheme
}

func (this *LoopContext) GetStatus() *status.Status {
	return this.status
}

func (this *LoopContext) GetEnvCache() env.EnvCache {
	return this.envCache
}
