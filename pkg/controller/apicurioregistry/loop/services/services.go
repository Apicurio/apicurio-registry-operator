package services

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/patcher"
)

type LoopServices struct {
	Clients  *client.Clients
	Patchers *patcher.Patchers

	KubeFactory       *factory.KubeFactory
	OcpFactory        *factory.OCPFactory
	MonitoringFactory *factory.MonitoringFactory
}

func NewLoopServices(ctx *context.LoopContext) *LoopServices {

	factoryKube := factory.NewKubeFactory(ctx)
	factoryOcp := factory.NewOCPFactory(ctx, factoryKube)
	factoryMonitoring := factory.NewMonitoringFactory(ctx, factoryKube)

	clients := client.NewClients(ctx)

	patchers := patcher.NewPatchers(ctx, clients)

	return &LoopServices{
		Clients:  clients,
		Patchers: patchers,

		KubeFactory:       factoryKube,
		OcpFactory:        factoryOcp,
		MonitoringFactory: factoryMonitoring,
	}
}
