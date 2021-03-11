package factory

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop/context"
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MonitoringFactory struct {
	ctx         *context.LoopContext
	kubeFactory *KubeFactory
}

func NewMonitoringFactory(ctx *context.LoopContext, kubeFactory *KubeFactory) *MonitoringFactory {
	return &MonitoringFactory{
		ctx:         ctx,
		kubeFactory: kubeFactory,
	}
}

func (this *MonitoringFactory) GetLabels() map[string]string {
	return this.kubeFactory.GetLabels()
}

func (this *MonitoringFactory) GetSelectorLabels() map[string]string {
	return this.kubeFactory.GetSelectorLabels()
}

func (this *MonitoringFactory) NewServiceMonitor(service *core.Service) *monitoring.ServiceMonitor {
	name := this.ctx.GetAppName().Str()
	namespace := this.ctx.GetAppNamespace().Str()

	return &monitoring.ServiceMonitor{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    this.GetLabels(),
		},
		Spec: monitoring.ServiceMonitorSpec{
			Endpoints: []monitoring.Endpoint{
				{
					Path:       "/metrics",
					TargetPort: &service.Spec.Ports[0].TargetPort,
				},
			},
			NamespaceSelector: monitoring.NamespaceSelector{
				MatchNames: []string{namespace},
			},
			Selector: meta.LabelSelector{
				MatchLabels: this.GetSelectorLabels(),
			},
		},
	}
}
