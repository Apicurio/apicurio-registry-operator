package context

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/go-logr/logr"
	"time"
)

var _ LoopContext = &LoopContextMock{}

type LoopContextMock struct {
	appName       common.Name
	appNamespace  common.Namespace
	log           logr.Logger
	resourceCache resources.ResourceCache
	envCache      env.EnvCache
	attempts      int
}

func NewLoopContextMock() *LoopContextMock {
	res := &LoopContextMock{
		appName:      common.Name("mock"),
		appNamespace: common.Namespace("mock"),
	}
	res.log = common.BuildLogger(true)
	res.resourceCache = resources.NewResourceCache()
	res.envCache = env.NewEnvCache(res.log)
	return res
}

func (this *LoopContextMock) GetLog() logr.Logger {
	return this.log
}

func (this *LoopContextMock) GetAppName() common.Name {
	return this.appName
}

func (this *LoopContextMock) GetAppNamespace() common.Namespace {
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

func (this *LoopContextMock) GetTestingSupport() *common.TestSupport {
	panic("Not implemented")
}
