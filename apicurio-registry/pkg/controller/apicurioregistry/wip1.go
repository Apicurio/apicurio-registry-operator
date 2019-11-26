package apicurioregistry

import (
	extensions "k8s.io/api/extensions/v1beta1"
)

func extractHost(serviceName string, ingress *extensions.Ingress) (*extensions.IngressRule, *extensions.HTTPIngressPath, *string) {
	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.ServiceName == serviceName {
				return &rule, &path, &rule.Host
			}
		}
	}
	return nil, nil, nil
}
