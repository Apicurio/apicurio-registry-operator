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
}

// Create a new context when the operator is deployed, provide mostly static data
func NewContext(c controller.Controller, scheme *runtime.Scheme, log logr.Logger, client client.Client) *Context {
	self := &Context{c: c, scheme: scheme, log: log, client: client}
	self.configuration = NewConfiguration(log)
	self.kubecl = NewKubeCl(self)
	self.patcher = NewPatcher(self)
	self.factory = NewFactory(self)
	return self
}

// Refresh context's state on each reconciliation loop execution,
// BEFORE CF execution
func (this *Context) update(spec *ar.ApicurioRegistry) {
	this.configuration.Update(spec)
}
