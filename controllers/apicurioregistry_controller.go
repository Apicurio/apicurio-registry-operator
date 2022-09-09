package controllers

import (
	go_ctx "context"
	"errors"
	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/cf"
	"github.com/Apicurio/apicurio-registry-operator/controllers/cf/condition"
	"github.com/Apicurio/apicurio-registry-operator/controllers/cf/kafkasql"
	"github.com/Apicurio/apicurio-registry-operator/controllers/client"
	c "github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/impl"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/resources"
	"github.com/go-logr/logr"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy "k8s.io/api/policy/v1beta1"
	cr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ApicurioRegistryReconciler{}

type ApicurioRegistryReconciler struct {
	log     logr.Logger
	clients *client.Clients
	testing *c.TestSupport
	loops   map[string]loop.ControlLoop
}

func NewApicurioRegistryReconciler(mgr manager.Manager, rootLog logr.Logger, testing *c.TestSupport) (*ApicurioRegistryReconciler, error) {

	clients := client.NewClients(
		rootLog.WithName("clients"),
		mgr.GetScheme(), mgr.GetConfig())

	isOCP, err := clients.Discovery().IsOCP()
	if err != nil {
		return nil, errors.New("could not determine cluster type")
	}
	if isOCP {
		rootLog.V(c.V_NORMAL).Info("This operator is running on OpenShift")
	} else {
		rootLog.V(c.V_NORMAL).Info("This operator is running on Kubernetes")
	}

	result := &ApicurioRegistryReconciler{
		log:     rootLog.WithName("controller"),
		clients: clients,
		testing: testing,
		loops:   make(map[string]loop.ControlLoop),
	}

	if err := result.setupWithManager(mgr); err != nil {
		return nil, err
	}

	return result, nil
}

func (this *ApicurioRegistryReconciler) setupWithManager(mgr cr.Manager) error {

	builder := cr.NewControllerManagedBy(mgr)

	builder.For(&ar.ApicurioRegistry{})

	builder.WithEventFilter(predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld.GetObjectKind().GroupVersionKind().Kind == "ApicurioRegistry" {
				// Ignore updates to the ApicurioRegistry status, in which case metadata.Generation does not change.
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			}
			return true
		},
	})

	builder.Owns(&apps.Deployment{})
	builder.Owns(&core.Service{})
	builder.Owns(&networking.Ingress{})
	builder.Owns(&policy.PodDisruptionBudget{})

	isMonitoring, err := this.clients.Discovery().IsMonitoringInstalled()
	if err != nil {
		return err
	}
	if isMonitoring {
		builder.Owns(&monitoring.ServiceMonitor{})
	} else {
		this.log.V(c.V_IMPORTANT).Info("Install prometheus-operator in your cluster to create ServiceMonitor objects, restart apicurio-registry operator after installing prometheus-operator")
	}

	return builder.Complete(this)
}

// Apicurio Registry CR
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries,verbs=*
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries/finalizers,verbs=update

// OpenShift
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=*

// Common
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=*
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=*
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
// +kubebuilder:rbac:groups=core,resources=pods;services;endpoints;persistentvolumeclaims;configmaps;secrets;services/finalizers,verbs=*
// +kubebuilder:rbac:groups=events,resources=events,verbs=*

// Monitoring
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=*

// Cluster Info (k8s vs. OCP)
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get

func (this *ApicurioRegistryReconciler) Reconcile(_ go_ctx.Context, request reconcile.Request) (reconcile.Result, error) {
	if this.testing.IsEnabled() {
		this.testing.ResetTimer()
	}

	appName := c.Name(request.Name)
	appNamespace := c.Namespace(request.Namespace)

	this.log.V(c.V_NORMAL).Info("reconciler executing")

	// Find the spec
	spec, err := this.clients.CRD().GetApicurioRegistry(appNamespace, appName)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Get the target control loop
	key := appNamespace.Str() + "/" + appName.Str() // TODO Use types.NamespacedName ?
	controlLoop, exists := this.loops[key]
	if exists {
		// If control loop exists, but spec is not found, do a cleanup
		if spec == nil {
			controlLoop.Cleanup()
			delete(this.loops, key)
			controlLoop.GetContext().GetLog().V(c.V_NORMAL).Info("context was deleted")
			return reconcile.Result{}, nil
		} else {
			// Run and reload spec into the cache
			controlLoop.GetContext().GetResourceCache().Set(resources.RC_KEY_SPEC, resources.NewResourceCacheEntry(appName, spec))
		}
	} else {
		if spec == nil {
			// Error
			return reconcile.Result{}, nil
		} else {
			// Create new loop, and requeue
			controlLoop = this.createNewLoop(appName, appNamespace)
			this.loops[key] = controlLoop
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// Loop is established, run it
	controlLoop.Run()

	// Reschedule if requested
	requeue, delay := controlLoop.GetContext().GetAndResetRequeue()
	return reconcile.Result{Requeue: requeue, RequeueAfter: delay}, nil
}

func (this *ApicurioRegistryReconciler) createNewLoop(appName c.Name, appNamespace c.Namespace) loop.ControlLoop {

	loopKey := appNamespace.Str() + "/" + appName.Str()
	log := this.log.WithValues("contextId", loopKey)
	log.V(c.V_NORMAL).Info("creating a new context")

	ctx := context.NewLoopContext(appName, appNamespace, log, this.clients, this.testing)
	loopServices := services.NewLoopServices(ctx)
	result := impl.NewControlLoopImpl(ctx, loopServices)

	//functions ordered so execution is optimized

	// Initialization, executed only once (or only for a short time)
	result.AddControlFunction(condition.NewInitializingCF(ctx, loopServices))
	result.AddControlFunction(cf.NewHostInitCF(ctx))

	//deployment
	result.AddControlFunction(cf.NewDeploymentCF(ctx, loopServices))

	//dependents of deployment
	result.AddControlFunction(cf.NewAffinityCF(ctx))
	result.AddControlFunction(cf.NewPodDisruptionBudgetCF(ctx, loopServices))
	result.AddControlFunction(cf.NewServiceMonitorCF(ctx, loopServices))
	result.AddControlFunction(cf.NewTolerationCF(ctx))
	result.AddControlFunction(cf.NewAnnotationsCF(ctx))

	//deployment modifiers
	result.AddControlFunction(cf.NewImageCF(ctx, loopServices))
	result.AddControlFunction(cf.NewImagePullPolicyCF(ctx))
	result.AddControlFunction(cf.NewImagePullSecretsCF(ctx))

	result.AddControlFunction(cf.NewReplicasCF(ctx, loopServices))

	//deployment env vars modifiers
	result.AddControlFunction(cf.NewSqlCF(ctx))
	result.AddControlFunction(kafkasql.NewKafkasqlCF(ctx))
	result.AddControlFunction(kafkasql.NewKafkasqlSecurityScramCF(ctx))
	result.AddControlFunction(kafkasql.NewKafkasqlSecurityTLSCF(ctx))
	result.AddControlFunction(cf.NewLogLevelCF(ctx))
	result.AddControlFunction(cf.NewProfileCF(ctx))
	result.AddControlFunction(cf.NewUICF(ctx))
	result.AddControlFunction(cf.NewKeycloakCF(ctx))

	//env vars from CR
	result.AddControlFunction(cf.NewEnvCF(ctx))
	//env vars applier
	result.AddControlFunction(cf.NewEnvApplyCF(ctx))

	//service
	result.AddControlFunction(cf.NewServiceCF(ctx, loopServices))

	//ingress (depends on service)
	result.AddControlFunction(cf.NewIngressCF(ctx, loopServices))

	//network policy
	result.AddControlFunction(cf.NewNetworkPolicyCF(ctx, loopServices))

	//dependents of ingress
	if isOCP, _ := this.clients.Discovery().IsOCP(); isOCP {
		result.AddControlFunction(cf.NewHostInitRouteOcpCF(ctx))
	}
	result.AddControlFunction(cf.NewHostCF(ctx, loopServices))

	// Other / Dependent on everything :)
	result.AddControlFunction(cf.NewLabelsCF(ctx, loopServices))
	result.AddControlFunction(condition.NewAppHealthCF(ctx, loopServices))

	return result
}
