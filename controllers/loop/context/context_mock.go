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

var _ LoopContext = &LoopContextMock{}

type LoopContextMock struct {
	appName           c.Name
	appNamespace      c.Namespace
	log               *zap.Logger
	resourceCache     resources.ResourceCache
	envCache          env.EnvCache
	attempts          int
	reconcileSequence int64
}

func NewLoopContextMock() *LoopContextMock {
	res := &LoopContextMock{
		appName:           c.Name("mock"),
		appNamespace:      c.Namespace("mock"),
		reconcileSequence: 0,
	}
	res.log = c.GetRootLogger(true)
	res.resourceCache = resources.NewResourceCache()
	res.envCache = env.NewEnvCache(res.log)
	return res
}

func (this *LoopContextMock) GetLog() *zap.Logger {
	return this.log
}

func (this *LoopContextMock) GetAppName() c.Name {
	return this.appName
}

func (this *LoopContextMock) GetAppNamespace() c.Namespace {
	return this.appNamespace
}

func (this *LoopContextMock) SetRequeueNow() {
	panic("not implemented")
}

func (this *LoopContextMock) SetRequeueDelaySoon() {
	panic("not implemented")
}

func (this *LoopContextMock) SetRequeueDelaySec(delay uint) {
	panic("not implemented")
}

func (this *LoopContextMock) Finalize() (bool, time.Duration) {
	if this.reconcileSequence == math.MaxInt64 {
		panic("int64 counter overflow. Restarting to reset.") // This will never happen
	}
	this.reconcileSequence += 1
	return false, 0
}

func (this *LoopContextMock) GetResourceCache() resources.ResourceCache {
	return this.resourceCache
}

func (this *LoopContextMock) GetClients() *client.Clients {
	panic("not implemented")
}

func (this *LoopContextMock) GetEnvCache() env.EnvCache {
	return this.envCache
}

func (this *LoopContextMock) SetAttempts(attempts int) {
	this.attempts = attempts
}

func (this *LoopContextMock) GetAttempts() int {
	return this.attempts
}

func (this *LoopContextMock) GetTestingSupport() *c.TestSupport {
	panic("not implemented")
}

func (this *LoopContextMock) GetSupportedFeatures() *c.SupportedFeatures {
	panic("not implemented")
}

func (this *LoopContextMock) GetReconcileSequence() int64 {
	return this.reconcileSequence
}
