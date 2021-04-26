package services

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/patcher"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status/conditions"
)

type LoopServices struct {
	clients  *client.Clients
	patchers *patcher.Patchers

	kubeFactory       *factory.KubeFactory
	monitoringFactory *factory.MonitoringFactory

	conditionManager conditions.ConditionManager
	status           *status.Status
}

func NewLoopServices(ctx *context.LoopContext) *LoopServices {
	this := &LoopServices{}
	this.kubeFactory = factory.NewKubeFactory(ctx)
	this.monitoringFactory = factory.NewMonitoringFactory(ctx, this.kubeFactory)
	this.clients = client.NewClients(ctx)
	this.conditionManager = conditions.NewConditionManager(ctx)
	this.status = status.NewStatus(ctx, this.conditionManager)
	this.patchers = patcher.NewPatchers(ctx, this.clients, this.kubeFactory, this.status)
	return this
}

func (this *LoopServices) BeforeRun() {
	this.patchers.Reload()
}

func (this *LoopServices) AfterRun() {
	this.conditionManager.AfterLoop() // TODO Unify nomenclature
	this.status.ComputeStatus()
	this.patchers.Execute()
}

func (this *LoopServices) GetClients() *client.Clients {
	return this.clients
}

func (this *LoopServices) GetPatchers() *patcher.Patchers {
	return this.patchers
}

func (this *LoopServices) GetKubeFactory() *factory.KubeFactory {
	return this.kubeFactory
}

func (this *LoopServices) GetMonitoringFactory() *factory.MonitoringFactory {
	return this.monitoringFactory
}

func (this *LoopServices) GetConditionManager() conditions.ConditionManager {
	return this.conditionManager
}

func (this *LoopServices) GetStatus() *status.Status {
	return this.status
}
