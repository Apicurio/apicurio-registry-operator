package apicurioregistry

import (
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// A long-lived singleton container for shared components
type Context struct {
	// More static stuff
	scheme *runtime.Scheme
	log    logr.Logger
	client client.Client
	c      controller.Controller

	// Components
	kubecl        *KubeCl
	configuration *Configuration
	factory       *Factory
	patcher       *Patcher

	controlFunctions []ControlFunction
}

// Create a new context when the operator is deployed, provide mostly static data
func NewContext(c controller.Controller, scheme *runtime.Scheme, log logr.Logger, client client.Client) *Context {
	self := &Context{c: c, scheme: scheme, log: log, client: client,
	//controlFunctions:make([]cf.ControlFunction)
	}
	self.configuration = NewConfiguration(log)
	self.kubecl = NewKubeCl(self)
	self.patcher = NewPatcher(self)
	self.factory = NewFactory(self)
	return self
}

func (this *Context) AddControlFunction(cf ControlFunction) {
	this.controlFunctions = append(this.controlFunctions, cf)
}



// Refresh context's state on each reconciliation loop execution,
// BEFORE CF execution
func (this *Context) Update(spec *ar.ApicurioRegistry) {
	this.configuration.Update(spec)
}


func (this *Context) GetControlFunctions() []ControlFunction {
	return this.controlFunctions
}


func (this *Context) GetLog() logr.Logger {
	return this.log
}

func (this *Context) GetKubeCl() *KubeCl {
	return this.kubecl
}

func (this *Context) GetConfiguration() *Configuration {
	return this.configuration
}

func (this *Context) GetPatcher() *Patcher {
	return this.patcher
}

func (this *Context) GetController() controller.Controller {
	return this.c
}

func (this *Context) GetFactory() *Factory {
	return this.factory
}

func (this *Context) GetScheme() *runtime.Scheme {
	return this.scheme
}

func (this *Context) GetClient() client.Client {
	return this.client
}