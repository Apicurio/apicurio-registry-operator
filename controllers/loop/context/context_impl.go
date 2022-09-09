package context

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/go-logr/logr"
	"time"
)

var _ LoopContext = &loopContext{}

// A long-lived singleton container for shared, data only, 0 dependencies, components
type loopContext struct {
	appName      common.Name
	appNamespace common.Namespace
	log          logr.Logger
	requeue      bool
	requeueDelay time.Duration

	resourceCache resources.ResourceCache
	envCache      env.EnvCache

	attempts int
	clients  *client.Clients

	testing *common.TestSupport
}

// Create a new context when the operator is deployed, provide mostly static data
func NewLoopContext(appName common.Name, appNamespace common.Namespace, log logr.Logger, clients *client.Clients, testing *common.TestSupport) LoopContext {
	this := &loopContext{
		appName:      appName,
		appNamespace: appNamespace,
		requeue:      false,
		requeueDelay: 0,
		clients:      clients,
		testing:      testing,
		log:          log,
	}
	//this.log = log.WithValues("app", appName.Str(), "namespace", appNamespace.Str())

	this.resourceCache = resources.NewResourceCache()

	this.envCache = env.NewEnvCache(log)

	return this
}

func (this *loopContext) GetLog() logr.Logger {
	return this.log
}

func (this *loopContext) GetAppName() common.Name {
	return this.appName
}

func (this *loopContext) GetAppNamespace() common.Namespace {
	return this.appNamespace
}

func (this *loopContext) SetRequeueNow() {
	this.SetRequeueDelaySec(0)
}

func (this *loopContext) SetRequeueDelaySoon() {
	this.SetRequeueDelaySec(5)
}

func (this *loopContext) SetRequeueDelaySec(delay uint) {
	d := time.Duration(delay) * time.Second
	if this.requeue == false || d < this.requeueDelay {
		this.requeueDelay = d
		this.requeue = true
	}
}

func (this *loopContext) GetAndResetRequeue() (bool, time.Duration) {
	defer func() {
		this.requeue = false
		this.requeueDelay = 0
	}()
	return this.requeue, this.requeueDelay
}

func (this *loopContext) GetClients() *client.Clients {
	return this.clients
}

func (this *loopContext) GetResourceCache() resources.ResourceCache {
	return this.resourceCache
}

func (this *loopContext) GetEnvCache() env.EnvCache {
	return this.envCache
}

func (this *loopContext) SetAttempts(attempts int) {
	this.attempts = attempts
}

func (this *loopContext) GetAttempts() int {
	return this.attempts
}

func (this *loopContext) GetTestingSupport() *common.TestSupport {
	return this.testing
}
