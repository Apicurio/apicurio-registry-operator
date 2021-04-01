module github.com/Apicurio/apicurio-registry-operator

go 1.14

require (
	github.com/go-logr/logr v0.3.0
	github.com/openshift/api v0.0.0-20210317213936-dcbf045ae1b8
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.46.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.46.0
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	sigs.k8s.io/controller-runtime v0.8.0
)
