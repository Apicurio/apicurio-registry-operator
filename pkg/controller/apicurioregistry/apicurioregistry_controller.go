package apicurioregistry

import (
	registry "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	ocp_apps "github.com/openshift/api/apps/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/kubernetes/pkg/apis/apps"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

	err = c.Watch(&source.Kind{Type: &registry.ApicurioRegistry{}}, &handler.EnqueueRequestForObject{}, predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to the ApicurioRegistry status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
	})
	if err != nil {
		return err
	}

	isocp, err := client.IsOCP()
	if err != nil {
		return err
	}
	if isocp {
		err := c.Watch(&source.Kind{Type: &ocp_apps.DeploymentConfig{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &registry.ApicurioRegistry{},
		})

		if err != nil {
			return err
		}
	} else {
		err = c.Watch(&source.Kind{Type: &apps.Deployment{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &registry.ApicurioRegistry{},
		})

		if err != nil {
			return err
		}
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &registry.ApicurioRegistry{},
	})

	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &extensions.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &registry.ApicurioRegistry{},
	})

	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &policy.PodDisruptionBudget{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &registry.ApicurioRegistry{},
	})

	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &monitoring.ServiceMonitor{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &registry.ApicurioRegistry{},
	})

	if err != nil {
		if api_errors.IsNotFound(err) {
			log.WithValues("type", "Warning", "reason", err.Error()).Info("Could not create ServiceMonitor watch. is prometheus-operator installed?")
		} else {
			return err
		}
	}

	return nil
}
