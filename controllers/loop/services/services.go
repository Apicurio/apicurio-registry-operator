package services

import (
	"github.com/Apicurio/apicurio-registry-operator/controllers/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/controllers/svc/patcher"
)

type LoopServices struct {
	clients  *client.Clients
	patchers *patcher.Patchers

	kubeFactory       *factory.KubeFactory
	monitoringFactory *factory.MonitoringFactory
}

func NewLoopServices(ctx *context.LoopContext) *LoopServices {

	factoryKube := factory.NewKubeFactory(ctx)
	factoryMonitoring := factory.NewMonitoringFactory(ctx, factoryKube)

	clients := client.NewClients(ctx)

	patchers := patcher.NewPatchers(ctx, clients, factoryKube)

	return &LoopServices{
		clients:  clients,
		patchers: patchers,

		kubeFactory:       factoryKube,
		monitoringFactory: factoryMonitoring,
	}
}

func (svcs *LoopServices) BeforeRun() {
	svcs.patchers.Reload()
}

func (svcs *LoopServices) AfterRun() {
	svcs.patchers.Execute()
}

func (svcs *LoopServices) GetClients() *client.Clients {
	return svcs.clients
}

func (svcs *LoopServices) GetPatchers() *patcher.Patchers {
	return svcs.patchers
}

func (svcs *LoopServices) GetKubeFactory() *factory.KubeFactory {
	return svcs.kubeFactory
}

func (svcs *LoopServices) GetMonitoringFactory() *factory.MonitoringFactory {
	return svcs.monitoringFactory
}
