package apicurioregistry

import (
	monitoring "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MonitoringFactory struct {
	ctx *Context
}

func NewMonitoringFactory(ctx *Context) *MonitoringFactory {
	return &MonitoringFactory{
		ctx: ctx,
	}
}

func (this *MonitoringFactory) NewServiceMonitor(service *core.Service) *monitoring.ServiceMonitor {
	name := this.ctx.configuration.GetAppName()
	namespace := this.ctx.configuration.GetAppNamespace()
	labels := make(map[string]string)
	for k, v := range service.ObjectMeta.Labels {
		labels[k] = v
	}

	return &monitoring.ServiceMonitor{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
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
				MatchLabels: labels,
			},
		},
	}
}
