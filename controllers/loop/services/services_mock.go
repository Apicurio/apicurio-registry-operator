package services

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/patcher"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status/conditions"
)

var _ LoopServices = &LoopServicesMock{}

type LoopServicesMock struct {
}

func NewLoopServicesMock(ctx context.LoopContext) *LoopServicesMock {
	this := &LoopServicesMock{}
	return this
}

func (this *LoopServicesMock) BeforeRun() {
	// NOOP
}

func (this *LoopServicesMock) AfterRun() {
	//this.conditionManager.AfterLoop() // TODO Unify nomenclature
	//this.status.ComputeStatus()
	//this.patchers.Execute()
}

func (this *LoopServicesMock) GetPatchers() *patcher.Patchers {
	panic("Not implemented")
}

func (this *LoopServicesMock) GetKubeFactory() *factory.KubeFactory {
	panic("Not implemented")
}

func (this *LoopServicesMock) GetMonitoringFactory() *factory.MonitoringFactory {
	panic("Not implemented")
}

func (this *LoopServicesMock) GetConditionManager() conditions.ConditionManager {
	panic("Not implemented")
}

func (this *LoopServicesMock) GetStatus() *status.Status {
	panic("Not implemented")
}
