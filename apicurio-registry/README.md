Apicurio Registry Operator
===

Requirements
---
* Docker
* [go](https://github.com/golang/go) (1.13+, with `export GO111MODULE='on'`), and `$GOPATH` and `$GOROOT` set. 
* [Operator SDK](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md) v0.9.0+    
* A running Kubernetes or [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) cluster, 
with `system:admin` access.

Build
---

Clone this repo under your `$GOPATH/src` dir and `cd` inside.

Pick a registry, e.g. `quay.io` and use `build.sh` script (or Operator SDK directly) 
to build the image:

`./build.sh build -r "$REGISTRY"`

And push it to the registry:

`./build.sh push -r "$REGISTRY"`

Installation
---

If you are testing on `minikube`, you can use the following commands 
to deploy and undeploy the operator, respectively, with an example CR:

`./build.sh mkdeploy -r "$REGISTRY"`

`./build.sh mkundeploy -r "$REGISTRY"`

Or you can perform the steps manually (see the `build.sh`):

1. Create resources and resource definitions on your cluster:
   
   `kubecl create -f deploy/service_account.yaml`
   
   `kubecl create -f deploy/role.yaml`
   
   `kubecl create -f deploy/role_binding.yaml`

1. Create operator CRD:
   
   `kubecl create -f deploy/crds/apicur_v1alpha1_apicurioregistry_crd.yaml`

1. Deploy the operator:

   `kubecl create -f deploy/operator.yaml`

1. Create an example deployment of Apicurio Registry (*mem*) using the operator:

   `kubecl create -f apicur_v1alpha1_apicurioregistry_cr.yaml`

Verify the deployment is active:

`kubecl get deployments`

`kubecl get pods`

(If the host is configured using `minikube ip` and `/etc/hosts` :)

`curl -v http://registry.example.com/health`

Development
---

Use the script to see the steps to build manually.

Apicurio Registry CRD
---

TODO

Note
---

This operator is in *alpha* stage, which means that while it's working and is able to replace the [registry templates](https://github.com/Apicurio/apicurio-registry/tree/master/distro/openshift-template),
some planned features are not implemented yet. 

It's been tested on `minikube`, but it has not been released on operator hub.
