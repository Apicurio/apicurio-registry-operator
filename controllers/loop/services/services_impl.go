package services

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/patcher"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status/conditions"
)

var _ LoopServices = &loopServices{}

type loopServices struct {
	patchers *patcher.Patchers

	kubeFactory       *factory.KubeFactory
	monitoringFactory *factory.MonitoringFactory

	conditionManager conditions.ConditionManager
	status           *status.Status
}

func NewLoopServices(ctx context.LoopContext) LoopServices {
	this := &loopServices{}
	this.kubeFactory = factory.NewKubeFactory(ctx)
	this.monitoringFactory = factory.NewMonitoringFactory(ctx, this.kubeFactory)
	this.conditionManager = conditions.NewConditionManager(ctx)
	this.status = status.NewStatus(ctx, this.conditionManager)
	this.patchers = patcher.NewPatchers(ctx, this.kubeFactory, this.status)
	return this
}

func (this *loopServices) BeforeRun() {
	this.patchers.Reload()
}

func (this *loopServices) AfterRun() {
	this.conditionManager.AfterLoop() // TODO Unify nomenclature
	this.status.ComputeStatus()
	this.patchers.Execute()
}

func (this *loopServices) GetPatchers() *patcher.Patchers {
	return this.patchers
}

func (this *loopServices) GetKubeFactory() *factory.KubeFactory {
	return this.kubeFactory
}

func (this *loopServices) GetMonitoringFactory() *factory.MonitoringFactory {
	return this.monitoringFactory
}

func (this *loopServices) GetConditionManager() conditions.ConditionManager {
	return this.conditionManager
}

func (this *loopServices) GetStatus() *status.Status {
	return this.status
}
