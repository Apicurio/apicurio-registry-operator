Apicurio Registry Operator
===

Requirements
---
* Docker
* [go](https://github.com/golang/go) ( `1.13+`, with `export GO111MODULE='on'`), and `$GOPATH` and `$GOROOT` set. 
* [Operator SDK](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md)    
* A running Kubernetes or [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) cluster, 
with `system:admin` access.

Build
---

Clone this repo under your `$GOPATH/src` dir and `cd` inside.

Pick a registry, e.g. `quay.io` and use Operator SDK to build the image:

`OPERATOR="${REGISTRY}/apicurio-registry-operator:v${VERSION}"`

`./script/init.sh -o $OPERATOR`

`operator-sdk build $OPERATOR`

`docker push $OPERATOR`

Installation
---

Create resources and resource definitions on your cluster:

`oc create -f deploy/service_account.yaml`

`oc create -f deploy/role.yaml`

`oc create -f deploy/role_binding.yaml`

Create operator CRD:

`oc create -f deploy/crds/registry_v1alpha1_apicurioregistry_crd.yaml`

Deploy the operator:

`oc create -f deploy/operator.yaml`

Create an example deployment of Apicurio Registry (*mem*) using the operator:

`oc create -f deploy/crds/registry_v1alpha1_apicurioregistry_cr.yaml`

Verify the deployment is active:

`oc get deployments`

`oc get pods`

Development
---

Run the following after updating the code:

`operator-sdk generate k8s`

`operator-sdk generate openapi`

Scripts
---

There are several scripts in the `./scripts` to quickly build or deploy the operator.
