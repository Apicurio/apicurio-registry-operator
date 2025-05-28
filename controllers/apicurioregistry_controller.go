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
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy_v1 "k8s.io/api/policy/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	cr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ApicurioRegistryReconciler{}

type ApicurioRegistryReconciler struct {
	log      *zap.Logger
	clients  *client.Clients
	testing  *c.TestSupport
	loops    map[string]loop.ControlLoop
	features *c.SupportedFeatures
}

func NewApicurioRegistryReconciler(mgr manager.Manager, rootLog *zap.Logger, testing *c.TestSupport) (*ApicurioRegistryReconciler, error) {

	clients := client.NewClients(
		rootLog.Named("clients"),
		mgr.GetScheme(), mgr.GetConfig())

	features := &c.SupportedFeatures{}

	isOCP, err := clients.Discovery().IsOCP()
	if err != nil {
		return nil, errors.New("could not determine cluster type")
	}
	features.IsOCP = isOCP
	if isOCP {
		rootLog.Sugar().Info("This operator is running on OpenShift")
	} else {
		rootLog.Sugar().Info("This operator is running on Kubernetes")
	}

	agi, err := clients.Discovery().GetVersionInfoForAPIGroup("policy")
	if err != nil {
		rootLog.Sugar().Errorw("could not determine supported API group versions for PodDisruptionBudget resource", "error", err)
		return nil, err
	}
	if _, found := c.FindString(agi.Versions, "v1"); found {
		features.SupportsPDBv1 = true
		rootLog.Info("API server supports PodDisruptionBudget v1")
	}
	if _, found := c.FindString(agi.Versions, "v1beta1"); found {
		features.SupportsPDBv1beta1 = true
		rootLog.Info("API server supports PodDisruptionBudget v1beta1")
	}
	features.PreferredPDBVersion = agi.PreferredVersion
	rootLog.Info("Preferred version of PodDisruptionBudget is " + agi.PreferredVersion)

	isMonitoring, err := clients.Discovery().IsMonitoringInstalled()
	if err != nil {
		rootLog.Sugar().Errorw("could not determine if monitoring resource is installed", "error", err)
		return nil, err
	}
	if !isMonitoring {
		rootLog.Sugar().Info("Install prometheus-operator in your cluster to create ServiceMonitor objects, restart apicurio-registry operator after installing prometheus-operator")
	}
	features.SupportsMonitoring = isMonitoring
	testing.SetSupportedFeatures(features)

	result := &ApicurioRegistryReconciler{
		log:      rootLog.Named("controller"),
		clients:  clients,
		testing:  testing,
		loops:    make(map[string]loop.ControlLoop),
		features: features,
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
	if this.features.SupportsPDBv1beta1 {
		builder.Owns(&policy_v1beta1.PodDisruptionBudget{})
	}
	if this.features.SupportsPDBv1 {
		builder.Owns(&policy_v1.PodDisruptionBudget{})
	}
	if this.features.SupportsMonitoring {
		builder.Owns(&monitoring.ServiceMonitor{})
	}

	return builder.Complete(this)
}

// Apicurio Registry CR
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries,verbs=*
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries/finalizers,verbs=update

// OpenShift
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=*
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=use

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

	appName := c.Name(request.Name)
	appNamespace := c.Namespace(request.Namespace)

	if this.testing.IsEnabled() {
		this.testing.ResetTimer(appNamespace.Str())
	}

	this.log.Sugar().Info("reconciler executing")

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
			controlLoop.GetContext().GetLog().Sugar().Info("context was deleted")
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
			controlLoop = this.createNewLoop(appName, appNamespace, this.features)
			this.loops[key] = controlLoop
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// Loop is established, run it
	controlLoop.Run()

	// Reschedule if requested
	requeue, delay := controlLoop.GetContext().Finalize()
	return reconcile.Result{Requeue: requeue, RequeueAfter: delay}, nil
}

func (this *ApicurioRegistryReconciler) createNewLoop(appName c.Name, appNamespace c.Namespace, features *c.SupportedFeatures) loop.ControlLoop {

	loopKey := appNamespace.Str() + "/" + appName.Str()
	log := this.log.Sugar().With("contextId", loopKey)
	log.Info("creating a new context")

	ctx := context.NewLoopContext(appName, appNamespace, log.Desugar(), this.clients, this.testing, features)
	loopServices := services.NewLoopServices(ctx)
	result := impl.NewControlLoopImpl(ctx, loopServices)

	//functions ordered so execution is optimized

	// Initialization, executed only once (or only for a short time)
	result.AddControlFunction(condition.NewInitializingCF(ctx, loopServices))
	result.AddControlFunction(cf.NewHostInitCF(ctx))

	//deployment
	result.AddControlFunction(cf.NewDeploymentCF(ctx, loopServices))

	//deployment modifiers
	result.AddControlFunction(cf.NewUpgradeCF(ctx))
	result.AddControlFunction(cf.NewPodTemplateSpecCF(ctx, loopServices))
	result.AddControlFunction(cf.NewAffinityCF(ctx))
	result.AddControlFunction(cf.NewTolerationCF(ctx))
	result.AddControlFunction(cf.NewAnnotationsCF(ctx))
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
	result.AddControlFunction(cf.NewCorsCF(ctx))

	//env vars from CR
	result.AddControlFunction(cf.NewEnvCF(ctx))
	//env vars applier
	result.AddControlFunction(cf.NewEnvApplyCF(ctx))

	//depends on deployment
	if features.SupportsPDBv1beta1 {
		result.AddControlFunction(cf.NewPodDisruptionBudgetV1beta1CF(ctx, loopServices))
	}
	if features.SupportsPDBv1 {
		result.AddControlFunction(cf.NewPodDisruptionBudgetV1CF(ctx, loopServices))
	}

	//service
	result.AddControlFunction(cf.NewServiceCF(ctx, loopServices))

	// service modifiers
	result.AddControlFunction(cf.NewHttpsCF(ctx, loopServices))

	// depends on service
	if features.SupportsMonitoring {
		// TODO Temporarily disabling this feature, needs improvements
		// mainly to support HTTPS
		//result.AddControlFunction(cf.NewServiceMonitorCF(ctx, loopServices))
	}

	// network policy
	result.AddControlFunction(cf.NewNetworkPolicyCF(ctx, loopServices))

	// ingress
	result.AddControlFunction(cf.NewIngressCF(ctx, loopServices))

	//dependents of ingress
	if features.IsOCP {
		result.AddControlFunction(cf.NewHostInitRouteOcpCF(ctx))
	}
	result.AddControlFunction(cf.NewHostCF(ctx, loopServices))
	result.AddControlFunction(cf.NewIngressAnnotationsCF(ctx))
	result.AddControlFunction(cf.NewIngressClassNameCF(ctx))

	// Other / Dependent on everything :)
	result.AddControlFunction(cf.NewLabelsCF(ctx, loopServices))
	result.AddControlFunction(condition.NewAppHealthCF(ctx, loopServices))

	return result
}
