package context

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"go.uber.org/zap"
	"time"
)

type LoopContext interface {
	GetLog() *zap.Logger
	GetAppName() c.Name
	GetAppNamespace() c.Namespace
	SetRequeueNow()
	SetRequeueDelaySoon()
	SetRequeueDelaySec(delay uint)
	Finalize() (bool, time.Duration)
	GetClients() *client.Clients
	GetResourceCache() resources.ResourceCache
	GetEnvCache() env.EnvCache
	SetAttempts(attempts int)
	GetAttempts() int
	GetTestingSupport() *c.TestSupport
	GetSupportedFeatures() *c.SupportedFeatures
	GetReconcileSequence() int64
}
