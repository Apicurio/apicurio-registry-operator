---
layout: default
title: Development
nav_order: 4
has_children: true
---

# Development

## Requirements

* Docker
* [go](https://github.com/golang/go) (1.13+, with `export GO111MODULE='on'`), and `$GOPATH` and `$GOROOT` set.
* [Operator SDK](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md) v11+
* A running Kubernetes, [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/),
  OpenShift, or Minishift deployment with system admin access.

TODO

## Build


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

## Installation

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
    $ cat ./deploy/cluster_role_binding.yaml | sed "s/{NAMESPACE}/$NAMESPACE/g" | kubectl apply -f -
    $ kubectl create -f ./deploy/cluster_role.yaml
    $ kubectl create -f ./deploy/cluster_role_binding.yaml
    ```

1. Create operator CRD:

    ```
    $ kubectl create -f ./deploy/crds/apicur_v1alpha1_apicurioregistry_crd.yaml
    ```

1. Deploy the operator:

    ```
    $ kubectl create -f ./deploy/operator.yaml
    ```

1. Create an example deployment of Apicurio Registry (in-memory) using the operator:

    ```
    $ kubectl create -f ./deploy/crds/apicur_v1alpha1_apicurioregistry_cr.yaml
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
