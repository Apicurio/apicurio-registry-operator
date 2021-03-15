module github.com/Apicurio/apicurio-registry-operator

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/coreos/prometheus-operator v0.38.0
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20191227052852-215e87163ea7 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/openshift/api v0.0.0-20200205133042-34f0ec8dab87
	github.com/openshift/client-go v0.0.0-20200116152001-92a2713fa240
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stoewer/go-strcase v1.2.0 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5 // indirect
	go.uber.org/goleak v1.1.10 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.19.0-alpha.1
	k8s.io/apiextensions-apiserver v0.19.0-alpha.1 // indirect
	k8s.io/apimachinery v0.19.0-alpha.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.19.2 // indirect
	k8s.io/gengo v0.0.0-20200413195148-3a45101e95ac // indirect
	k8s.io/klog/v2 v2.2.0 // indirect
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.9 // indirect
	sigs.k8s.io/controller-runtime v0.6.5
	sigs.k8s.io/structured-merge-diff/v4 v4.0.1 // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.34.0
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
	k8s.io/code-generator => github.com/openshift/kubernetes-code-generator v0.0.0-20191216140939-db549faca3fe
)
