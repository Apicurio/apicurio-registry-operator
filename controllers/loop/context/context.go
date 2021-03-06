package context

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	sigs_client "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// A long-lived singleton container for shared, data only, 0 dependencies, components
type LoopContext struct {
	appName      common.Name
	appNamespace common.Namespace
	log          logr.Logger
	requeue      bool
	requeueDelay time.Duration

	client sigs_client.Client
	scheme *runtime.Scheme

	resourceCache resources.ResourceCache

	envCache env.EnvCache

	attempts int
}

// Create a new context when the operator is deployed, provide mostly static data
func NewLoopContext(appName common.Name, appNamespace common.Namespace, log logr.Logger, scheme *runtime.Scheme, client sigs_client.Client) *LoopContext {
	this := &LoopContext{
		appName:      appName,
		appNamespace: appNamespace,
		requeue:      false,
		requeueDelay: 0,
	}
	this.log = log.WithValues("app", appName.Str(), "namespace", appNamespace.Str())

	this.client = client
	this.scheme = scheme

	this.resourceCache = resources.NewResourceCache()

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

func (this *LoopContext) SetRequeueNow() {
	this.SetRequeueDelaySec(0)
}

func (this *LoopContext) SetRequeueDelaySoon() {
	this.SetRequeueDelaySec(5)
}

func (this *LoopContext) SetRequeueDelaySec(delay uint) {
	d := time.Duration(delay) * time.Second
	if this.requeue == false || d < this.requeueDelay {
		this.requeueDelay = d
		this.requeue = true
	}
}

func (this *LoopContext) GetAndResetRequeue() (bool, time.Duration) {
	defer func() {
		this.requeue = false
		this.requeueDelay = 0
	}()
	return this.requeue, this.requeueDelay
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

func (this *LoopContext) GetEnvCache() env.EnvCache {
	return this.envCache
}

func (this *LoopContext) SetAttempts(attempts int) {
	this.attempts = attempts
}

func (this *LoopContext) GetAttempts() int {
	return this.attempts
}
