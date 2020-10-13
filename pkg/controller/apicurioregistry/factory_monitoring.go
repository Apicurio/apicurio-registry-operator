package apicurioregistry

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MonitoringFactory struct {
	ctx loop.ControlLoopContext
}

func NewMonitoringFactory(ctx loop.ControlLoopContext) *MonitoringFactory {
	return &MonitoringFactory{
		ctx: ctx,
	}
}

func (this *MonitoringFactory) GetLabels() map[string]string {
	return this.ctx.RequireService(svc.SVC_KUBE_FACTORY).(KubeFactory).GetLabels()
}

func (this *MonitoringFactory) GetSelectorLabels() map[string]string {
	return this.ctx.RequireService(svc.SVC_KUBE_FACTORY).(KubeFactory).GetSelectorLabels()
}

func (this *MonitoringFactory) NewServiceMonitor(service *core.Service) *monitoring.ServiceMonitor {
	name := this.ctx.GetAppName()
	namespace := this.ctx.GetAppNamespace()

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
