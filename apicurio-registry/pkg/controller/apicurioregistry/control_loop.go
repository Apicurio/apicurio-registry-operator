package apicurioregistry

import (
	"context"
	"errors"
	ar "github.com/apicurio/apicurio-operators/apicurio-registry/pkg/apis/apicur/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ApicurioRegistryReconciler{}

type ApicurioRegistryReconciler struct {
	client           client.Client
	scheme           *runtime.Scheme
	controlFunctions []ControlFunction
	notInitialized   bool
	ctx              *Context
	controller       controller.Controller
}

func NewApicurioRegistryReconciler(mgr manager.Manager) *ApicurioRegistryReconciler {

	return &ApicurioRegistryReconciler{
		client:           mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		controlFunctions: []ControlFunction{},
		notInitialized:   true,
	}
}

func (this *ApicurioRegistryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("ApicurioRegistryReconciler executing.")

	// GetConfig the spec
	spec := &ar.ApicurioRegistry{}
	err := this.client.Get(context.TODO(), request.NamespacedName, spec)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Init
	if this.notInitialized {
		this.ctx = NewContext(this.controller, this.scheme, reqLogger, this.client)

		var cf ControlFunction
		cf = NewDeploymentCF(this.ctx)
		this.AddControlFunction(cf)

		cf = NewServiceCF(this.ctx)
		this.AddControlFunction(cf)

		cf = NewIngressCF(this.ctx)
		this.AddControlFunction(cf)

		cf = NewImageConfigCF(this.ctx)
		this.AddControlFunction(cf)

		cf = NewConfReplicasCF(this.ctx)
		this.AddControlFunction(cf)

		cf = NewHostConfigCF(this.ctx)
		this.AddControlFunction(cf)

		cf = NewEnvCF(this.ctx)
		this.AddControlFunction(cf)

		this.notInitialized = false
	}

	// Context update
	this.ctx.update(spec)

	// GetConfig possible config errors
	if errs := this.ctx.configuration.GetErrors(); len(*errs) > 0 {
		for _, v := range *errs {
			err := errors.New(v)
			log.Error(err, v)
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// The LOOP
	requeue := false
	for _, v := range this.controlFunctions {
		err = v.Sense(spec, request)
		if err != nil {
			log.Error(err, "Error during the SENSE phase of '"+v.Describe()+"' CF.")
			requeue = true
			continue
		}
		var discrepancy bool
		discrepancy, err = v.Compare(spec)
		if err != nil {
			log.Error(err, "Error during the COMPARE phase of '"+v.Describe()+"' CF.")
			requeue = true
			continue
		}
		if !discrepancy {
			continue
		}
		var changed bool
		changed, err = v.Respond(spec)
		if changed {
			requeue = true
		}
		if err != nil {
			log.Error(err, "Error during the RESPOND phase of '"+v.Describe()+"' CF.")
			requeue = true
			continue
		}
	}

	// Update the status
	spec = this.ctx.factory.CreateSpec(spec)
	err = this.client.Status().Update(context.TODO(), spec)
	if err != nil {
		log.Error(err, "Error updating status")
		return reconcile.Result{}, err
	}

	// Run patcher
	this.ctx.patcher.Execute()

	// TODO should we return errors or rely on panic to signal a critical failure?
	return reconcile.Result{Requeue: requeue}, nil // err
}

func (this *ApicurioRegistryReconciler) AddControlFunction(cf ControlFunction) {
	this.controlFunctions = append(this.controlFunctions, cf)
}

func (this *ApicurioRegistryReconciler) setController(c controller.Controller) {
	this.controller = c
}
