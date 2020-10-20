package apicurioregistry

import (
	"context"
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/cf"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/common"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/impl"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
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
	loops      map[loop.ControlLoop]int
}

func NewApicurioRegistryReconciler(mgr manager.Manager) *ApicurioRegistryReconciler {

	return &ApicurioRegistryReconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		loops:  make(map[loop.ControlLoop]int),
	}
}

func (this *ApicurioRegistryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	appName := common.Name(request.Name)
	appNamespace := common.Namespace(request.Namespace)

	log.Info("Reconciler executing.")

	// =====

	// Find the spec
	spec, err := this.getApicurioRegistryResource(appNamespace, appName)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Get the existing control loop
	var controlLoop loop.ControlLoop = nil
	for c := range this.loops {
		if c.GetContext().GetAppName() == appName {
			controlLoop = c
			break
		}
	}

	// If control loop exists, but spec is not found, do a cleanup
	if controlLoop != nil && spec == nil {
		controlLoop.Cleanup()
		delete(this.loops, controlLoop)
		controlLoop.GetContext().GetLog().Info("Context was deleted.")
		return reconcile.Result{}, nil
	}

	// If control loop was not found and spec exists, create a new one
	if controlLoop == nil && spec != nil {
		controlLoop = this.createNewLoop(appName, appNamespace)
	}

	if controlLoop == nil || spec == nil {
		return reconcile.Result{}, nil
	}

	// =======
	// Context is established

	controlLoop.Run()

	// ======
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
	ctx := impl.NewDefaultContext(appName, appNamespace, this.controller, this.scheme, log, this.client)
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
		c.AddControlFunction(cf.NewStatusCF(ctx))

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
		c.AddControlFunction(cf.NewStatusCF(ctx))
		c.AddControlFunction(cf.NewStreamsCF(ctx))

		c.AddControlFunction(cf.NewStreamsSecurityScramCF(ctx))
		c.AddControlFunction(cf.NewStreamsSecurityTLSCF(ctx))
		c.AddControlFunction(cf.NewTolerationCF(ctx))
		c.AddControlFunction(cf.NewUICF(ctx))
	}

	return c
}
