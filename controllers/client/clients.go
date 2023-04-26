// Some code in this file was adopted from https://github.com/atlasmap/atlasmap-operator
package client

import (
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type Clients struct {
	log *zap.Logger
	//ctx context.LoopContext
	//config           *rest.Config
	kubeClient       *KubeClient
	ocpClient        *OCPClient
	crdClient        *CRDClient
	monitoringClient *MonitoringClient
	discoveryClient  *DiscoveryClient
	scheme           *runtime.Scheme
}

func NewClients(log *zap.Logger, scheme *runtime.Scheme, config *rest.Config) *Clients {
	this := &Clients{
		scheme: scheme,
		log:    log,
	}
	//config, err := inClusterConfig()
	//if err != nil {
	//	common.Fatal(ctx.GetLog(), err, "Could not configure clients.")
	//}
	//this.config = config
	//config := ctx.GetClientConfig()
	log.Sugar().Debugw("client config values", "config", config)

	this.kubeClient = NewKubeClient(log, scheme, config)

	this.ocpClient = NewOCPClient(log, scheme, config)

	this.crdClient = NewCRDClient(log, scheme, config)

	this.monitoringClient = NewMonitoringClient(log, scheme, config)

	this.discoveryClient = NewDiscoveryClient(log, config)

	return this
}

func (this *Clients) OCP() *OCPClient {
	return this.ocpClient
}

func (this *Clients) Kube() *KubeClient {
	return this.kubeClient
}

func (this *Clients) CRD() *CRDClient {
	return this.crdClient
}

func (this *Clients) Monitoring() *MonitoringClient {
	return this.monitoringClient
}

func (this *Clients) Discovery() *DiscoveryClient {
	return this.discoveryClient
}
