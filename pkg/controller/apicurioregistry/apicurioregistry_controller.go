package apicurioregistry

import (
	registry "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_apicurioregistry")

func Add(mgr manager.Manager) error {

	r := NewApicurioRegistryReconciler(mgr)

	// Create a new controller
	c, err := controller.New("apicurioregistry-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	r.setController(c)

	err = c.Watch(&source.Kind{Type: &registry.ApicurioRegistry{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}
