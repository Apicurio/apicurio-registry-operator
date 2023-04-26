package context

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	"time"
)

var _ LoopContext = &LoopContextMock{}

type LoopContextMock struct {
	appName       c.Name
	appNamespace  c.Namespace
	log           *zap.Logger
	resourceCache resources.ResourceCache
	envCache      env.EnvCache
	attempts      int
}

func NewLoopContextMock() *LoopContextMock {
	res := &LoopContextMock{
		appName:      c.Name("mock"),
		appNamespace: c.Namespace("mock"),
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
	panic("Not implemented")
}

func (this *LoopContextMock) SetRequeueDelaySoon() {
	panic("Not implemented")
}

func (this *LoopContextMock) SetRequeueDelaySec(delay uint) {
	panic("Not implemented")
}

func (this *LoopContextMock) GetAndResetRequeue() (bool, time.Duration) {
	panic("Not implemented")
}

func (this *LoopContextMock) GetResourceCache() resources.ResourceCache {
	return this.resourceCache
}

func (this *LoopContextMock) GetClients() *client.Clients {
	panic("Not implemented")
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
	panic("Not implemented")
}

func (this *LoopContextMock) GetSupportedFeatures() *c.SupportedFeatures {
	panic("Not implemented")
}
