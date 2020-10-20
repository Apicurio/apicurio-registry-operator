package loop

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type ControlLoopContext interface {
	GetLog() logr.Logger

	GetAppName() common.Name

	GetAppNamespace() common.Namespace

	GetService(name string) (interface{}, bool)

	RequireService(name string) interface{}

	GetController() controller.Controller

	GetScheme() *runtime.Scheme

	SetRequeue()

	GetAndResetRequeue() bool
}
