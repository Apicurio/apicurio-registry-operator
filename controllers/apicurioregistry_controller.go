package controllers

import (
	ctx "context"

	"github.com/Apicurio/apicurio-registry-operator/controllers/cf/condition"

	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/client"
	"github.com/go-logr/logr"
	ocp_apps "github.com/openshift/api/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	policy "k8s.io/api/policy/v1beta1"

	ar "github.com/Apicurio/apicurio-registry-operator/api/v1"
	"github.com/Apicurio/apicurio-registry-operator/controllers/cf"
	"github.com/Apicurio/apicurio-registry-operator/controllers/cf/kafkasql"
	"github.com/Apicurio/apicurio-registry-operator/controllers/common"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop"
	loop_context "github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/impl"
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/services"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	sigs_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ApicurioRegistryReconciler{}

type ApicurioRegistryReconciler struct {
	client sigs_client.Client
	scheme *runtime.Scheme
	//controller controller.Controller
	loops map[string]loop.ControlLoop
	log   logr.Logger
}

func NewApicurioRegistryReconciler(mgr manager.Manager, rootLog logr.Logger) (*ApicurioRegistryReconciler, error) {

	r := &ApicurioRegistryReconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		loops:  make(map[string]loop.ControlLoop),
		log:    rootLog.WithName("controllers").WithValues("controller", "ApicurioRegistry"),
	}
	if err := r.setupWithManager(mgr); err != nil {
		return nil, err
	}

	return r, nil
}

// Apicurio Registry CR
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries,verbs=*
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=registry.apicur.io,resources=apicurioregistries/finalizers,verbs=update

// OpenShift
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=*
// +kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs,verbs=*

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

func (this *ApicurioRegistryReconciler) Reconcile(reconcileCtx ctx.Context /* TODO or context.TODO()*/, request reconcile.Request) (reconcile.Result, error) {
	appName := common.Name(request.Name)
	appNamespace := common.Namespace(request.Namespace)

	this.log.Info("Reconciler executing.")

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
	requeue, delay := controlLoop.GetContext().GetAndResetRequeue()
	return reconcile.Result{Requeue: requeue, RequeueAfter: delay}, nil
}

//func (this *ApicurioRegistryReconciler) setController(c controller.Controller) {
//	this.controller = c
//}

// Returns nil if the resource is not found, but request was OK
func (this *ApicurioRegistryReconciler) getApicurioRegistryResource(appNamespace common.Namespace, appName common.Name) (*ar.ApicurioRegistry, error) {
	specList := &ar.ApicurioRegistryList{}
	listOps := sigs_client.ListOptions{Namespace: appNamespace.Str()}
	err := this.client.List(ctx.TODO(), specList, &listOps)
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

	this.log.Info("Creating new context")
	ctx := loop_context.NewLoopContext(appName, appNamespace, this.log, this.scheme, this.client)
	loopServices := services.NewLoopServices(ctx)
	c := impl.NewControlLoopImpl(ctx, loopServices)

	isOCP, _ := client.IsOCP()
	if isOCP {
		this.log.Info("This operator is running on OpenShift")
	} else {
		this.log.Info("This operator is running on Kubernetes")
	}

	//functions ordered so execution is optimized

	// Initialization, executed only once (or only for a short time)
	c.AddControlFunction(condition.NewInitializingCF(ctx, loopServices))
	c.AddControlFunction(cf.NewHostInitCF(ctx))

	//deployment
	c.AddControlFunction(cf.NewDeploymentCF(ctx, loopServices))

	//dependents of deployment
	c.AddControlFunction(cf.NewAffinityCF(ctx))
	c.AddControlFunction(cf.NewPodDisruptionBudgetCF(ctx, loopServices))
	c.AddControlFunction(cf.NewServiceMonitorCF(ctx, loopServices))
	c.AddControlFunction(cf.NewTolerationCF(ctx))
	c.AddControlFunction(cf.NewAnnotationsCF(ctx))

	//deployment modifiers
	c.AddControlFunction(cf.NewImageCF(ctx, loopServices))
	c.AddControlFunction(cf.NewImagePullPolicyCF(ctx))
	c.AddControlFunction(cf.NewImagePullSecretsCF(ctx))

	c.AddControlFunction(cf.NewReplicasCF(ctx, loopServices))

	//deployment env vars modifiers
	c.AddControlFunction(cf.NewSqlCF(ctx))
	c.AddControlFunction(kafkasql.NewKafkasqlCF(ctx))
	c.AddControlFunction(kafkasql.NewKafkasqlSecurityScramCF(ctx))
	c.AddControlFunction(kafkasql.NewKafkasqlSecurityTLSCF(ctx))
	c.AddControlFunction(cf.NewLogLevelCF(ctx))
	c.AddControlFunction(cf.NewProfileCF(ctx))
	c.AddControlFunction(cf.NewUICF(ctx))
	c.AddControlFunction(cf.NewKeycloakCF(ctx))

	//env vars applier
	c.AddControlFunction(cf.NewEnvCF(ctx))

	//service
	c.AddControlFunction(cf.NewServiceCF(ctx, loopServices))

	//ingress (depends on service)
	c.AddControlFunction(cf.NewIngressCF(ctx, loopServices))

	//network policy
	c.AddControlFunction(cf.NewNetworkPolicyCF(ctx, loopServices))

	//dependents of ingress
	if isOCP {
		c.AddControlFunction(cf.NewHostInitRouteOcpCF(ctx))
	}
	c.AddControlFunction(cf.NewHostCF(ctx, loopServices))

	// Other / Dependent on everything :)
	c.AddControlFunction(cf.NewLabelsCF(ctx, loopServices))
	c.AddControlFunction(condition.NewAppHealthCF(ctx, loopServices))

	return c
}

// ######################################

func (this *ApicurioRegistryReconciler) setupWithManager(mgr ctrl.Manager) error {

	builder := ctrl.NewControllerManagedBy(mgr).
		Named("ApicurioRegistry-controller")

	builder.For(&ar.ApicurioRegistry{}).WithEventFilter(predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to the ApicurioRegistry status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	})

	isOCP, err := client.IsOCP()
	if err != nil {
		return err
	}
	if isOCP {
		builder.Owns(&ocp_apps.DeploymentConfig{})
	} else {
		builder.Owns(&apps.Deployment{})
	}

	builder.Owns(&corev1.Service{})
	builder.Owns(&networking.Ingress{})
	builder.Owns(&policy.PodDisruptionBudget{})

	isMonitoring, err := client.IsMonitoringInstalled()
	if err != nil {
		return err
	}
	if isMonitoring {
		builder.Owns(&monitoring.ServiceMonitor{})
	} else {
		this.log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects, restart apicurio-registry operator after installing prometheus-operator")
	}

	return builder.Complete(this)
}
