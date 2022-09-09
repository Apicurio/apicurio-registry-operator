package services

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/patcher"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/status/conditions"
)

type LoopServices interface {
	BeforeRun()
	AfterRun()
	GetPatchers() *patcher.Patchers
	GetKubeFactory() *factory.KubeFactory
	GetMonitoringFactory() *factory.MonitoringFactory
	GetConditionManager() conditions.ConditionManager
	GetStatus() *status.Status
}
