package apicurioregistry

import (
	"context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/cf"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/impl"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/configuration"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/patcher"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
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
	listOps := sigs_client.ListOptions{Namespace: appNamespace}
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
	loop.GetContext().RequireService(svc.SVC_CONFIGURATION).(*configuration.Configuration).Update(spec)
	specEntry0 := resources.NewResourceCacheEntry(spec.Name, spec)
	loop.GetContext().RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Set(resources.RC_KEY_SPEC, specEntry0)
	loop.GetContext().RequireService(svc.SVC_PATCHERS).(*patcher.Patchers).Reload()
    // =====
	
	loop.Run()

	// This has to be done explicitly ATM, TODO Add status CF, use the current `configuration` as status cache (+error handling)
	specEntry, _ := loop.GetContext().RequireService(svc.SVC_RESOURCE_CACHE).(resources.ResourceCache).Get(resources.RC_KEY_SPEC)
	specEntry.ApplyPatch(func(value interface{}) interface{} {
		spec := value.(*ar.ApicurioRegistry).DeepCopy()
		spec.Status = *loop.GetContext().RequireService(svc.SVC_KUBE_FACTORY).(*factory.KubeFactory).CreateStatus(spec)
		return spec
	})

	// ======
	// Create or patch resources in resource cache
	loop.GetContext().RequireService(svc.SVC_PATCHERS).(*patcher.Patchers).Execute()

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

	isOCP, _ := ctx.RequireService(svc.SVC_CLIENTS).(*client.Clients).IsOCP()
	if isOCP {
		log.Info("This operator is running on OpenShift")

		// Keep alphabetical!
		c.AddControlFunction(cf.NewAffinityOcpCF(ctx))
		c.AddControlFunction(cf.NewDeploymentOcpCF(ctx))
		c.AddControlFunction(cf.NewEnvOcpCF(ctx))
		c.AddControlFunction(cf.NewHostCF(ctx))
		c.AddControlFunction(cf.NewHostInitCF(ctx))

		c.AddControlFunction(cf.NewHostInitRouteOcpCF(ctx))
		c.AddControlFunction(cf.NewImageOcpCF(ctx))
		c.AddControlFunction(cf.NewInfinispanCF(ctx))
		c.AddControlFunction(cf.NewIngressCF(ctx))
		c.AddControlFunction(cf.NewJpaCF(ctx))

		c.AddControlFunction(cf.NewKafkaCF(ctx))
		c.AddControlFunction(cf.NewLabelsOcpCF(ctx))
		c.AddControlFunction(cf.NewLogLevelCF(ctx))
		c.AddControlFunction(cf.NewOperatorPodCF(ctx))
		c.AddControlFunction(cf.NewPodDisruptionBudgetCF(ctx))

		c.AddControlFunction(cf.NewProfileCF(ctx))
		c.AddControlFunction(cf.NewReplicasOcpCF(ctx))
		c.AddControlFunction(cf.NewServiceCF(ctx))
		c.AddControlFunction(cf.NewServiceMonitorCF(ctx))
		c.AddControlFunction(cf.NewStreamsCF(ctx))

		c.AddControlFunction(cf.NewStreamsSecurityScramOcpCF(ctx))
		c.AddControlFunction(cf.NewStreamsSecurityTLSOcpCF(ctx))
		c.AddControlFunction(cf.NewTolerationOcpCF(ctx))
		c.AddControlFunction(cf.NewUICF(ctx))

	} else {
		log.Info("This operator is running on Kubernetes")

		// Keep alphabetical!
		c.AddControlFunction(cf.NewAffinityCF(ctx))
		c.AddControlFunction(cf.NewDeploymentCF(ctx))
		c.AddControlFunction(cf.NewEnvCF(ctx))
		c.AddControlFunction(cf.NewHostCF(ctx))
		c.AddControlFunction(cf.NewHostInitCF(ctx))

		c.AddControlFunction(cf.NewImageCF(ctx))
		c.AddControlFunction(cf.NewInfinispanCF(ctx))
		c.AddControlFunction(cf.NewIngressCF(ctx))
		c.AddControlFunction(cf.NewJpaCF(ctx))
		c.AddControlFunction(cf.NewKafkaCF(ctx))

		c.AddControlFunction(cf.NewLabelsCF(ctx))
		c.AddControlFunction(cf.NewLogLevelCF(ctx))
		c.AddControlFunction(cf.NewOperatorPodCF(ctx))
		c.AddControlFunction(cf.NewPodDisruptionBudgetCF(ctx))
		c.AddControlFunction(cf.NewProfileCF(ctx))

		c.AddControlFunction(cf.NewReplicasCF(ctx))
		c.AddControlFunction(cf.NewServiceCF(ctx))
		c.AddControlFunction(cf.NewServiceMonitorCF(ctx))
		c.AddControlFunction(cf.NewStreamsCF(ctx))
		c.AddControlFunction(cf.NewStreamsSecurityScramCF(ctx))

		c.AddControlFunction(cf.NewStreamsSecurityTLSCF(ctx))
		c.AddControlFunction(cf.NewTolerationCF(ctx))
		c.AddControlFunction(cf.NewUICF(ctx))
	}

	return c
}
