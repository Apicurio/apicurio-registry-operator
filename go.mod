module github.com/Apicurio/apicurio-registry-operator

go 1.13

require (
	github.com/coreos/prometheus-operator v0.38.0
	github.com/go-logr/logr v0.1.0
	github.com/openshift/api v0.0.0-20200205133042-34f0ec8dab87
	github.com/openshift/client-go v0.0.0-20200116152001-92a2713fa240
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.34.0
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
	k8s.io/code-generator => github.com/openshift/kubernetes-code-generator v0.0.0-20191216140939-db549faca3fe
)
