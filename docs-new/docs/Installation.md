---
layout: default
title: Installation
nav_order: 2
#has_children: true
---

# Installation
{: .no_toc }

There are two main options when installing Apicurio Registry Operator.

Note: Information about deployment of Apicurio Registry itself is in the [Configuration]({{ site.baseurl }}{% link docs/Configuration.md %}) page. 

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}



## Operator Lifecycle Manager

You can use [Operator Lifecycle Manager (OLM)](https://docs.openshift.com/container-platform/latest/operators/understanding_olm/olm-understanding-olm.html) to manage the operator deployment for you. Apicurio Registry Operator is released to [Operator Hub](https://operatorhub.io/?keyword=apicurio+registry), so if your cluster has OLM istalled, you should be able to use the web UI to deploy the Operator. This is the preferred installation option for users.

[Development documentation]({{ site.baseurl }}{% link docs/Development.md %}) explains how to configure OLM on your cluster to deploy development versions of the Operator.

## Command Line Interface

You can use CLI (`kubectl` or `oc`). To deploy an operator, you have to create several types of resources in your cluster:

  - *Service Account*, *Role*, and *Cluster Role* - to configure permissions for the operator.
  - *Custom Resource Definition (CRD)* - which defines resources that the operator understands, and are created by users to provide configuration for Apicurio Registry deployment.
  - *Deployment* - to deploy the operator pod.

These resources are located in the `deploy` and `deploy/crds` directories.

We have packaged the resources into a single file so you can deploy the latest development operator version using a single command:

```bash
export NAMESPACE="default" &&
export GIT_REF="master" &&
curl -sSL "https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/$GIT_REF/docs/install/install.yaml" | sed "s/{NAMESPACE}/$NAMESPACE/g" | kubectl apply -f -
```

You can modify the variables to deploy to a different namespace, or use a released version of the operator by selecting a git tag (0.0.3+).

(If you are deploying to OpenShift, you can use `oc` with the same arguments.)

If you want to deploy the resources
Use following commands to deploy the resources:

```bash
export NAMESPACE="<target namespace>"
kubectl create -f ./deploy/service_account.yaml
kubectl create -f ./deploy/role.yaml
kubectl create -f ./deploy/role_binding.yaml
kubectl create -f ./deploy/cluster_role.yaml
cat ./deploy/cluster_role_binding.yaml | sed "s/{NAMESPACE}/$NAMESPACE/g" | kubectl apply -f -
kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_crd.yaml
kubectl create -f ./deploy/operator.yaml
```

Now, you can create an `ApicurioRegistry` CR to instruct the operator to deploy Apicurio Registry. [Configuration page]({{ site.baseurl }}{% link docs/Configuration.md %}) explains what this CD looks like.

---

Visit [Development Documentation]({{ site.baseurl }}{% link docs/Development.md %}) for more details on building and deploying development versions of the operator.

<!---

TODO

operator-sdk run --local

-->



