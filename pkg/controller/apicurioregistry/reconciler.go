package apicurioregistry

import (
	"context"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/cf"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	loop_context "github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/impl"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	sigs_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ApicurioRegistryReconciler{}

type ApicurioRegistryReconciler struct {
	client     sigs_client.Client
	scheme     *runtime.Scheme
	controller controller.Controller
	loops      map[string]loop.ControlLoop
}

func NewApicurioRegistryReconciler(mgr manager.Manager) *ApicurioRegistryReconciler {

	return &ApicurioRegistryReconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		loops:  make(map[string]loop.ControlLoop),
	}
}

func (this *ApicurioRegistryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	appName := common.Name(request.Name)
	appNamespace := common.Namespace(request.Namespace)

	log.Info("Reconciler executing.")

	// Find the spec
	spec, err := this.getApicurioRegistryResource(appNamespace, appName)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Get the target control loop
	key := appNamespace.Str() + "/" + appName.Str()
	controlLoop, exists := this.loops[key]
	if exists {
		// If control loop exists, but spec is not found, do a cleanup
		if spec == nil {
			controlLoop.Cleanup()
			delete(this.loops, key)
			controlLoop.GetContext().GetLog().Info("Context was deleted.")
			return reconcile.Result{}, nil
		} // else OK, run
	} else {
		if spec == nil {
			// Error
			return reconcile.Result{}, nil
		} else {
			// Create new loop, and run
			controlLoop = this.createNewLoop(appName, appNamespace)
			this.loops[key] = controlLoop
		}
	}

	// Loop is established, run it
	controlLoop.Run()

	// Reschedule if requested
	return reconcile.Result{Requeue: controlLoop.GetContext().GetAndResetRequeue()}, nil
}

func (this *ApicurioRegistryReconciler) setController(c controller.Controller) {
	this.controller = c
}

// Returns nil if the resource is not found, but request was OK
func (this *ApicurioRegistryReconciler) getApicurioRegistryResource(appNamespace common.Namespace, appName common.Name) (*ar.ApicurioRegistry, error) {
	specList := &ar.ApicurioRegistryList{}
	listOps := sigs_client.ListOptions{Namespace: appNamespace.Str()}
	err := this.client.List(context.TODO(), specList, &listOps)
	if err != nil {
		if api_errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var spec *ar.ApicurioRegistry = nil

	for i, specItem := range specList.Items {
		if common.Name(specItem.Name) == appName && common.Namespace(specItem.Namespace) == appNamespace {
			spec = &specList.Items[i]
		}
	}

	return spec, nil
}

func (this *ApicurioRegistryReconciler) createNewLoop(appName common.Name, appNamespace common.Namespace) loop.ControlLoop {

	log.Info("Creating new context")
	ctx := loop_context.NewLoopContext(appName, appNamespace, log, this.scheme, this.client)
	services := services.NewLoopServices(ctx)
	c := impl.NewControlLoopImpl(ctx, services)

	isOCP, _ := client.IsOCP()
	if isOCP {
		log.Info("This operator is running on OpenShift")
	} else {
		log.Info("This operator is running on Kubernetes")
	}

	//functions ordered so execution is optimized

	//host init, executed only once
	c.AddControlFunction(cf.NewHostInitCF(ctx))

	//deployment
	c.AddControlFunction(cf.NewDeploymentCF(ctx, services))

	//dependents of deployment
	c.AddControlFunction(cf.NewAffinityCF(ctx))
	c.AddControlFunction(cf.NewPodDisruptionBudgetCF(ctx, services))
	c.AddControlFunction(cf.NewServiceMonitorCF(ctx, services))
	c.AddControlFunction(cf.NewTolerationCF(ctx))
	//deployment modifiers
	c.AddControlFunction(cf.NewImageCF(ctx))
	c.AddControlFunction(cf.NewReplicasCF(ctx))
	//deployment env vars modifiers
	c.AddControlFunction(cf.NewInfinispanCF(ctx))
	c.AddControlFunction(cf.NewSqlCF(ctx))
	c.AddControlFunction(cf.NewStreamsCF(ctx))
	c.AddControlFunction(cf.NewStreamsSecurityScramCF(ctx))
	c.AddControlFunction(cf.NewStreamsSecurityTLSCF(ctx))
	c.AddControlFunction(cf.NewLogLevelCF(ctx))
	c.AddControlFunction(cf.NewProfileCF(ctx))
	c.AddControlFunction(cf.NewUICF(ctx))
	//env vars applier
	c.AddControlFunction(cf.NewEnvCF(ctx))

	//service
	c.AddControlFunction(cf.NewServiceCF(ctx, services))

	//ingress (depends on service)
	c.AddControlFunction(cf.NewIngressCF(ctx, services))

	//dependents of ingress
	if isOCP {
		c.AddControlFunction(cf.NewHostInitRouteOcpCF(ctx))
	}
	c.AddControlFunction(cf.NewHostCF(ctx))

	//dependent on everything :)
	c.AddControlFunction(cf.NewLabelsCF(ctx, services))

	return c
}
