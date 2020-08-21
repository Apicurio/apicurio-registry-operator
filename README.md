Apicurio Registry Operator
===

Requirements
---
* Docker
* [go](https://github.com/golang/go) (1.13+, with `export GO111MODULE='on'`), and `$GOPATH` and `$GOROOT` set. 
* [Operator SDK](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md) v11+    
* A running Kubernetes, [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/), 
  OpenShift, or Minishift deployment with system admin access.

Quickstart
---

Choose your `$NAMESPACE` and you can deploy the latest published operator version using a single command:
```
$ curl -sSL https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/master/docs/resources/install.yaml | sed "s/{NAMESPACE}/$NAMESPACE/g" | kubectl apply -f -
```

(If you are deploying to OpenShift, you can use `oc` with the same arguments.)
 
This will deploy the latest development version of the operator from the `master` branch, 
but you can deploy other versions as well. 
Use a different branch or tag, or edit the operator image reference in the file. 

To create a new Apicurio Registry deployment, the fastest way is to use in-memory persistence option and one of the example CRs:
```
$ kubectl create -f https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/master/docs/resources/example-cr/in-memory.yaml
```
Note: The in-memory deployment is not suitable for production. We recommend using Kafka Streams persistence option for that.
See the contents of `./docs` for more information.

Build
---

Clone this repo under your `$GOPATH/src/github.com/Apicurio` dir and `cd` inside.

Pick a registry, e.g. [quay.io](quay.io)/user and use `build.sh` script (or Operator SDK directly) 
to build the image:

```
$ ./build.sh build -r "$REGISTRY"
```

And push it to the registry:

```
$ ./build.sh push -r "$REGISTRY"
```

Installation
---

If you are testing on Minikube, you can use the following commands 
to deploy and undeploy the operator, respectively, with an example CR:

```
$ ./build.sh mkdeploy -r "$REGISTRY"
$ ./build.sh mkundeploy -r "$REGISTRY"
```

Or you can perform the steps manually (see the `build.sh`):

1. Create resources and resource definitions on your cluster (choose your $NAMESPACE):
   
    ```
    $ kubectl create -f ./deploy/service_account.yaml
    $ kubectl create -f ./deploy/role.yaml
    $ kubectl create -f ./deploy/role_binding.yaml
    $ cat ./deploy/cluster_role_binding.yaml | sed "s/{NAMESPACE}/$NAMESPACE/g" | kubectl apply -f -
    $ kubectl create -f ./deploy/cluster_role.yaml
    ```

1. Create operator CRD:
   
    ```
    $ kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_crd.yaml
    ```

1. Deploy the operator:

    ```
    $ kubectl create -f ./deploy/operator.yaml
    ```

1. Create an example deployment of Apicurio Registry (in-memory) using the operator:

    ```
    $ kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_cr.yaml
    ```

1. Verify that the deployment is active:

    ```
    $ kubectl get deployments
    $ kubectl get pods
    ```
   
1. Make an HTTP request:
    
    ```
    $ curl -H "Host: registry.example.com" http://$(minikube ip)/health
    ```

    You can also configure the host using `minikube ip` and `/etc/hosts`: 

    ```
    $ curl -v http://registry.example.com/health
    ```

Notes
---

This operator is still in development. Please create an issue on GitHub if you come across any problems.

You can find more documentation in `./docs`.
