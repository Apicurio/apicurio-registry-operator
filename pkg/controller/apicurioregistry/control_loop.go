package apicurioregistry

import (
	"context"
	"reflect"
	"strconv"

	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	oappsv1 "github.com/openshift/api/apps/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	contexts   map[string]*Context
}

func NewApicurioRegistryReconciler(mgr manager.Manager) *ApicurioRegistryReconciler {

	return &ApicurioRegistryReconciler{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		contexts: make(map[string]*Context),
	}
}

func (this *ApicurioRegistryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	app := request.Name

	reqLogger := log.WithValues("namespace", request.Namespace, "app", app)
	reqLogger.Info("Reconciler executing.")


	instance := &ar.ApicurioRegistry{}
	err := this.client.Get(context.TODO(), request.NamespacedName, instance)

	if err != nil {
		if api_errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("No Custom Resource found named %s. Checking for dependent objects to delete."+ request.Name)
			instance.ObjectMeta = metav1.ObjectMeta{
				Name:      request.Name,
				Namespace: request.Namespace,
			}
			deployed, err := getDeployedResources(instance,this.client)
			if err != nil {
				return reconcile.Result{}, err
			}
			_, err = this.cleanResources(instance, nil, deployed)
			return reconcile.Result{}, err
		}

	}

	// =====
	// Get all apicurio registry CRs, in order to select or create the control loop context.

	specList := &ar.ApicurioRegistryList{}
	listOps := client.ListOptions{Namespace: request.Namespace}
	err = this.client.List(context.TODO(), specList, &listOps)
	if err != nil {
		if api_errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	var spec *ar.ApicurioRegistry = nil

	for i, specItem := range specList.Items {
		key := specItem.Name
		_, ok := this.contexts[key]
		if !ok {
			this.contexts[key] = this.createNewContext(key)
		}
		if app == key {
			spec = &specList.Items[i] // Note: Do not use spec = &specItem
		}
	}

	if spec == nil {
		_, ok := this.contexts[app]
		if ok {
			delete(this.contexts, app)
			reqLogger.Info("Context was deleted.")
		}
		return reconcile.Result{}, nil
	}

	ctx := this.contexts[app]

	// =======
	// Context is established

	// Context update
	ctx.Update(spec)
	ctx.GetPatchers().Reload()

	// CONTROL LOOP
	maxAttempts := len(ctx.GetControlFunctions()) * 2
	attempt := 0
	for ; attempt < maxAttempts; attempt++ {
		ctx.GetLog().WithValues("attempt", strconv.Itoa(attempt), "maxAttempts", strconv.Itoa(maxAttempts)).
			Info("Control loop executing.")
		// Run the CFs until we exceed the limit or the state has stabilized,
		// i.e. no action was taken by any CF
		stabilized := true
		for _, cf := range ctx.GetControlFunctions() {
			cf.Sense()
			discrepancy := cf.Compare()
			if discrepancy {
				ctx.GetLog().WithValues("cf", cf.Describe()).Info("Control function responding.")
				cf.Respond()
				stabilized = false
				break // Loop is restarted as soon as an action was taken
			}
		}

		// This has to be done explicitly ATM, TODO Add status CF, use the current `configuration` as status cache (+error handling)
		specEntry, _ := ctx.GetResourceCache().Get(RC_KEY_SPEC)
		specEntry.ApplyPatch(func(value interface{}) interface{} {
			spec := value.(*ar.ApicurioRegistry).DeepCopy()
			spec.Status = *ctx.GetKubeFactory().CreateStatus(spec)
			return spec
		})

		if stabilized {
			ctx.GetLog().Info("Control loop is stable.")
			break
		}
	}
	if attempt == maxAttempts {
		panic("Control loop stabilization limit exceeded.")
	}

	// ======
	// Create or patch resources in resource cache
	ctx.GetPatchers().Execute()

	// ======
	return reconcile.Result{Requeue: ctx.GetAndResetRequeue()}, nil
}

func (this *ApicurioRegistryReconciler) setController(c controller.Controller) {
	this.controller = c
}

// Create a new context for the given ApicurioRegistry CR.
// Choose te CFs based on the environment (currently Kubernetes vs. OpenShift)
func (this *ApicurioRegistryReconciler) createNewContext(appName string) *Context {

	log.Info("Creating new context")
	c := NewContext(this.controller, this.scheme, log.WithValues("app", appName), this.client)

	isOCP, _ := c.GetClients().IsOCP()
	if isOCP {
		log.Info("This operator is running on OpenShift")

		// Keep alphabetical!
		c.AddControlFunction(NewAffinityOcpCF(c))
		c.AddControlFunction(NewDeploymentOcpCF(c))
		c.AddControlFunction(NewEnvOcpCF(c))
		c.AddControlFunction(NewHostCF(c))
		c.AddControlFunction(NewHostInitCF(c))
		c.AddControlFunction(NewHostInitRouteOcpCF(c))

		c.AddControlFunction(NewImageOcpCF(c))
		c.AddControlFunction(NewInfinispanCF(c))
		c.AddControlFunction(NewIngressCF(c))
		c.AddControlFunction(NewJpaCF(c))
		c.AddControlFunction(NewKafkaCF(c))

		c.AddControlFunction(NewLogLevelCF(c))
		c.AddControlFunction(NewPodDisruptionBudgetCF(c))
		c.AddControlFunction(NewProfileCF(c))
		c.AddControlFunction(NewReplicasOcpCF(c))
		c.AddControlFunction(NewServiceCF(c))
		c.AddControlFunction(NewStreamsCF(c))

		c.AddControlFunction(NewStreamsSecurityScramOcpCF(c))
		c.AddControlFunction(NewStreamsSecurityTLSOcpCF(c))
		c.AddControlFunction(NewTolerationOcpCF(c))
		c.AddControlFunction(NewUICF(c))

	} else {
		log.Info("This operator is running on Kubernetes")

		// Keep alphabetical!
		c.AddControlFunction(NewAffinityCF(c))
		c.AddControlFunction(NewDeploymentCF(c))
		c.AddControlFunction(NewEnvCF(c))
		c.AddControlFunction(NewHostCF(c))
		c.AddControlFunction(NewHostInitCF(c))
		c.AddControlFunction(NewImageCF(c))

		c.AddControlFunction(NewInfinispanCF(c))
		c.AddControlFunction(NewIngressCF(c))
		c.AddControlFunction(NewJpaCF(c))
		c.AddControlFunction(NewKafkaCF(c))
		c.AddControlFunction(NewLogLevelCF(c))

		c.AddControlFunction(NewPodDisruptionBudgetCF(c))
		c.AddControlFunction(NewProfileCF(c))
		c.AddControlFunction(NewReplicasCF(c))
		c.AddControlFunction(NewServiceCF(c))
		c.AddControlFunction(NewStreamsCF(c))
		c.AddControlFunction(NewStreamsSecurityScramCF(c))

		c.AddControlFunction(NewStreamsSecurityTLSCF(c))
		c.AddControlFunction(NewTolerationCF(c))
		c.AddControlFunction(NewUICF(c))
	}

	return c
}

func (reconciler *ApicurioRegistryReconciler) cleanResources(instance *ar.ApicurioRegistry, requestedResources []resource.KubernetesResource, deployed map[reflect.Type][]resource.KubernetesResource) (bool, error) {
	reqLogger := log.WithValues("namespace", instance.Namespace, "app", instance.Name)

	writer := write.New(reconciler.client).WithOwnerController(instance, reconciler.scheme)
	//Compare what's deployed with what should be deployed
	requested := compare.NewMapBuilder().Add(requestedResources...).ResourceMap()
	comparator := compare.NewMapComparator()
	deltas := comparator.Compare(deployed, requested)

	var hasUpdates bool
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		reqLogger.Info("", "instances of ", resourceType, "Will delete", len(delta.Removed))

		removed, err := writer.RemoveResources(delta.Removed)
		if err != nil {
			return false, err
		}
		hasUpdates = hasUpdates || removed
	}
	return hasUpdates, nil
}



func getDeployedResources(instance *ar.ApicurioRegistry, client client.Client) (map[reflect.Type][]resource.KubernetesResource, error) {
	reqLogger := log.WithValues("namespace", instance.Namespace, "app", instance.Name)

	reader := read.New(client).WithNamespace(instance.Namespace)
	resourceMap, err := reader.ListAll(
		&oappsv1.DeploymentConfigList{},
		//&routev1.RouteList{},

	)
	if err != nil {
		reqLogger.Error(err, "Failed to list deployed objects. ", err)
		return nil, err
	}
	return resourceMap, nil
}

