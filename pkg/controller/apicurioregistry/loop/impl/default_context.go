package impl

import (
	base "github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

var _ loop.ControlLoopContext = &defaultContext{}

// A long-lived singleton container for shared components
type defaultContext struct {
	appName      string
	appNamespace string
	log          logr.Logger
	requeue      bool
	services     map[string]interface{}
}

// Create a new context when the operator is deployed, provide mostly static data
func NewDefaultContext(appName string, appNamespace string, c controller.Controller, scheme *runtime.Scheme, log logr.Logger, client client.Client) *defaultContext {
	this := &defaultContext{
		appName:      appName,
		appNamespace: appNamespace,
		log:          log,
		requeue:      false,
		services:     make(map[string]interface{}, 16),
	}
	this.services[svc.SVC_CONTROLLER] = c
	this.services[svc.SVC_SCHEME] = scheme
	this.services[svc.SVC_NATIVE_CLIENT] = client

	this.services[svc.SVC_CONFIGURATION] = base.NewConfiguration(log)

	this.services[svc.SVC_CLIENTS] = base.NewClients(this)
	this.services[svc.SVC_PATCHERS] = base.NewPatchers(this)

	this.services[svc.SVC_KUBE_FACTORY] = base.NewKubeFactory(this)
	this.services[svc.SVC_OCP_FACTORY] = base.NewOCPFactory(this)

	this.services[svc.SVC_RESOURCE_CACHE] = base.NewResourceCache()
	this.services[svc.SVC_ENV_CACHE] = base.NewEnvCache()

	return this
}

func (this *defaultContext) GetLog() logr.Logger {
	return this.log
}

func (this *defaultContext) GetAppName() string {
	return this.appName
}

func (this *defaultContext) GetAppNamespace() string {
	return this.appNamespace
}

func (this *defaultContext) GetService(name string) (interface{}, bool) {
	service, exists := this.services[name]
	return service, exists
}

func (this *defaultContext) RequireService(name string) interface{} {
	service, exists := this.GetService(name)
	if !exists {
		panic("Could not provide service " + name)
	}
	return service
}

func (this *defaultContext) GetController() controller.Controller {
	return this.RequireService(svc.SVC_CONTROLLER).(controller.Controller)
}

func (this *defaultContext) GetScheme() *runtime.Scheme {
	return this.RequireService(svc.SVC_SCHEME).(*runtime.Scheme)
}

func (this *defaultContext) SetRequeue() {
	this.requeue = true
}

func (this *defaultContext) GetAndResetRequeue() bool {
	res := this.requeue
	this.requeue = false
	return res
}
