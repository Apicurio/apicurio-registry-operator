package context

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	"math"
	"time"
)

var _ LoopContext = &loopContext{}

// A long-lived singleton container for shared, data only, 0 dependencies, components
type loopContext struct {
	appName           c.Name
	appNamespace      c.Namespace
	log               *zap.Logger
	requeue           bool
	requeueDelay      time.Duration
	resourceCache     resources.ResourceCache
	envCache          env.EnvCache
	attempts          int
	clients           *client.Clients
	testing           *c.TestSupport
	features          *c.SupportedFeatures
	reconcileSequence int64
}

// Create a new context when the operator is deployed, provide mostly static data
func NewLoopContext(appName c.Name, appNamespace c.Namespace, log *zap.Logger, clients *client.Clients, testing *c.TestSupport, features *c.SupportedFeatures) LoopContext {
	this := &loopContext{
		appName:           appName,
		appNamespace:      appNamespace,
		requeue:           false,
		requeueDelay:      0,
		clients:           clients,
		testing:           testing,
		log:               log,
		features:          features,
		reconcileSequence: 0,
	}
	this.resourceCache = resources.NewResourceCache()
	this.envCache = env.NewEnvCache(log)
	return this
}

func (this *loopContext) GetLog() *zap.Logger {
	return this.log
}

func (this *loopContext) GetAppName() c.Name {
	return this.appName
}

func (this *loopContext) GetAppNamespace() c.Namespace {
	return this.appNamespace
}

func (this *loopContext) SetRequeueNow() {
	this.SetRequeueDelaySec(0)
}

func (this *loopContext) SetRequeueDelaySoon() {
	this.SetRequeueDelaySec(5)
}

func (this *loopContext) SetRequeueDelaySec(delay uint) {
	this.log.Sugar().Debugln("SetRequeueDelaySec called with", delay)
	d := time.Duration(delay) * time.Second
	if this.requeue == false || d < this.requeueDelay {
		this.requeueDelay = d
		this.requeue = true
	}
}

func (this *loopContext) Finalize() (bool, time.Duration) {
	defer func() {
		this.requeue = false
		this.requeueDelay = 0
	}()
	if this.reconcileSequence == math.MaxInt64 {
		panic("int64 counter overflow. Restarting to reset.") // This will never happen
	}
	this.reconcileSequence += 1
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

func (this *loopContext) GetTestingSupport() *c.TestSupport {
	return this.testing
}

func (this *loopContext) GetSupportedFeatures() *c.SupportedFeatures {
	return this.features
}

func (this *loopContext) GetReconcileSequence() int64 {
	return this.reconcileSequence
}
