package apicurioregistry

import (
	"context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/impl"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ApicurioRegistryReconciler{}

type ApicurioRegistryReconciler struct {
	client     client.Client
	scheme     *runtime.Scheme
	controller controller.Controller
	//contexts   map[string]*Context

	loops map[loop.ControlLoop]int
}

func NewApicurioRegistryReconciler(mgr manager.Manager) *ApicurioRegistryReconciler {

	return &ApicurioRegistryReconciler{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		//contexts: make(map[string]*Context),

		loops: make(map[loop.ControlLoop]int),
	}
}

func (this *ApicurioRegistryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	appName := request.Name
	appNamespace := request.Namespace

	log.Info("Reconciler executing.")

	// =====

	// Find the spec
	specList := &ar.ApicurioRegistryList{}
	listOps := client.ListOptions{Namespace: appNamespace}
	err := this.client.List(context.TODO(), specList, &listOps)
	if err != nil {
		if api_errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	var spec *ar.ApicurioRegistry = nil

	for i, specItem := range specList.Items {
		if appName == specItem.Name {
			spec = &specList.Items[i] // Note: Do not use spec = &specItem
		}
	}

	// Get the relevant loop
	var loop loop.ControlLoop = nil
	for c := range this.loops {
		if c.GetContext().GetAppName() == appName {
			loop = c
			break
		}
	}

	if loop == nil && spec == nil {
		// error
		return reconcile.Result{}, nil
	}

	// If loop exists but spec not found, perform a cleanup
	if loop != nil && spec == nil {
		loop.Cleanup()
		delete(this.loops, loop)
		loop.GetContext().GetLog().Info("Context was deleted.")
		return reconcile.Result{}, nil
	}

	// If empty create it
	if loop == nil && spec != nil {
		loop = this.createNewLoop(appName, appNamespace)
	}

	// =======
	// Context is established

	// Context update
	loop.GetContext().RequireService(svc.SVC_CONFIGURATION).(Configuration).Update(spec)
	specEntry0 := NewResourceCacheEntry(spec.Name, spec)
	loop.GetContext().RequireService(svc.SVC_RESOURCE_CACHE).(ResourceCache).Set(RC_KEY_SPEC, specEntry0)
	loop.GetContext().RequireService(svc.SVC_PATCHERS).(Patchers).Reload()
    // =====
	
	loop.Run()

	// This has to be done explicitly ATM, TODO Add status CF, use the current `configuration` as status cache (+error handling)
	specEntry, _ := loop.GetContext().RequireService(svc.SVC_RESOURCE_CACHE).(ResourceCache).Get(RC_KEY_SPEC)
	specEntry.ApplyPatch(func(value interface{}) interface{} {
		spec := value.(*ar.ApicurioRegistry).DeepCopy()
		spec.Status = *loop.GetContext().RequireService(svc.SVC_KUBE_FACTORY).(KubeFactory).CreateStatus(spec)
		return spec
	})

	// ======
	// Create or patch resources in resource cache
	loop.GetContext().RequireService(svc.SVC_PATCHERS).(Patchers).Execute()

	// ======
	return reconcile.Result{Requeue: loop.GetContext().GetAndResetRequeue()}, nil
}

func (this *ApicurioRegistryReconciler) setController(c controller.Controller) {
	this.controller = c
}

func (this *ApicurioRegistryReconciler) getApicurioRegistryList(appName string) *loop.ControlLoop {
	return nil
}

func (this *ApicurioRegistryReconciler) createNewLoop(appName string, appNamespace string) loop.ControlLoop {

	log.Info("Creating new context")
	ctx := impl.NewDefaultContext(appName, appNamespace, this.controller, this.scheme, log.WithValues("app", appName), this.client)
    c := impl.NewControlLoopImpl(ctx)

	isOCP, _ := ctx.RequireService(svc.SVC_CLIENTS).(Clients).IsOCP()
	if isOCP {
		log.Info("This operator is running on OpenShift")

		// Keep alphabetical!
		c.AddControlFunction(NewAffinityOcpCF(ctx))
		c.AddControlFunction(NewDeploymentOcpCF(ctx))
		c.AddControlFunction(NewEnvOcpCF(ctx))
		c.AddControlFunction(NewHostCF(ctx))
		c.AddControlFunction(NewHostInitCF(ctx))

		c.AddControlFunction(NewHostInitRouteOcpCF(ctx))
		c.AddControlFunction(NewImageOcpCF(ctx))
		c.AddControlFunction(NewInfinispanCF(ctx))
		c.AddControlFunction(NewIngressCF(ctx))
		c.AddControlFunction(NewJpaCF(ctx))

		c.AddControlFunction(NewKafkaCF(ctx))
		c.AddControlFunction(NewLabelsOcpCF(ctx))
		c.AddControlFunction(NewLogLevelCF(ctx))
		c.AddControlFunction(NewOperatorPodCF(ctx))
		c.AddControlFunction(NewPodDisruptionBudgetCF(ctx))

		c.AddControlFunction(NewProfileCF(ctx))
		c.AddControlFunction(NewReplicasOcpCF(ctx))
		c.AddControlFunction(NewServiceCF(ctx))
		c.AddControlFunction(NewServiceMonitorCF(ctx))
		c.AddControlFunction(NewStreamsCF(ctx))

		c.AddControlFunction(NewStreamsSecurityScramOcpCF(ctx))
		c.AddControlFunction(NewStreamsSecurityTLSOcpCF(ctx))
		c.AddControlFunction(NewTolerationOcpCF(ctx))
		c.AddControlFunction(NewUICF(ctx))

	} else {
		log.Info("This operator is running on Kubernetes")

		// Keep alphabetical!
		c.AddControlFunction(NewAffinityCF(ctx))
		c.AddControlFunction(NewDeploymentCF(ctx))
		c.AddControlFunction(NewEnvCF(ctx))
		c.AddControlFunction(NewHostCF(ctx))
		c.AddControlFunction(NewHostInitCF(ctx))

		c.AddControlFunction(NewImageCF(ctx))
		c.AddControlFunction(NewInfinispanCF(ctx))
		c.AddControlFunction(NewIngressCF(ctx))
		c.AddControlFunction(NewJpaCF(ctx))
		c.AddControlFunction(NewKafkaCF(ctx))

		c.AddControlFunction(NewLabelsCF(ctx))
		c.AddControlFunction(NewLogLevelCF(ctx))
		c.AddControlFunction(NewOperatorPodCF(ctx))
		c.AddControlFunction(NewPodDisruptionBudgetCF(ctx))
		c.AddControlFunction(NewProfileCF(ctx))

		c.AddControlFunction(NewReplicasCF(ctx))
		c.AddControlFunction(NewServiceCF(ctx))
		c.AddControlFunction(NewServiceMonitorCF(ctx))
		c.AddControlFunction(NewStreamsCF(ctx))
		c.AddControlFunction(NewStreamsSecurityScramCF(ctx))

		c.AddControlFunction(NewStreamsSecurityTLSCF(ctx))
		c.AddControlFunction(NewTolerationCF(ctx))
		c.AddControlFunction(NewUICF(ctx))
	}

	return c
}
