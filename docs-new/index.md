---
layout: default
title: Home
nav_order: 1
description: "Apicurio Registry Operator Documentation"
permalink: /
nav_order: 1
---

# Apicurio Registry Operator
{: .no_toc }


Welcome to the user and developer documentation for Apicurio Registry Operator.
{: .fs-6 .fw-300 }


[Get Started](#quickstart){: .btn .btn-green .fs-5 .mb-4 .mb-md-0 .mr-2 } [View it on GitHub](https://github.com/Apicurio/apicurio-registry-operator){: .btn .fs-5 .mb-4 .mb-md-0 }

<br/>

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Apicurio Registry

Apicurio Registry stores and retrieves API designs and event schemas, 
and gives you control of their evolution.

For more information about the Operand, visit [apicur.io](https://www.apicur.io/registry/) 

## Apicurio Registry Operator

Provides a quick and easy way to deploy and manage an Apicurio Registry on Kubernetes or OpenShift.

The Operator supports basic installation and configuration of the Registry, 
so you can access the API and the UI in a few minutes.

<!-- 

For more information about features, visit [Releases](#releases) 

-->

## Prerequisites

Following platforms are supported:

|Platform|Required Version|
|---|---|
|Kubernetes|1.12+|
|OpenShift|3.11.200+, 4+|

The operator does not deploy and manage storage for the Registry yet. 
Therefore, some persistence options require that the chosen service is already set up.

## Quickstart

### Operator

There are several installation options, but the simplest one requires executing a single command.

Choose the `$NAMESPACE` to use:

```
export NAMESPACE="default"
```

and decide if you want to deploy the latest *released* version:

```
curl -sSL https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/master/docs/resources/install.yaml | 
sed "s/{NAMESPACE}/$NAMESPACE/g" | kubectl apply -f -
```

or the latest *development* build from the master branch:

```
curl -sSL https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/master/docs/resources/install-dev.yaml | 
sed "s/{NAMESPACE}/$NAMESPACE/g" | kubectl apply -f -
```

*Note: If you are deploying to OpenShift, use `oc` with the same arguments.*

You can find more information in the [Installation](docs/Installation) page.

### Apicurio Registry

To create a new Apicurio Registry deployment, the fastest way is to use in-memory persistence option and one of the example CRs:

`kubectl create -f https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/master/docs/resources/example-cr/in-memory.yaml`

You can use the in-memory persistence option to try out the Registry quickly,
but is not suitable for production. 
We recommend using Kafka Streams storage option for that.

Find more information in the [Configuration](docs/Configuration) page.

## Get Help

This operator is still in development. Any suggestions, issue reports an pull requests are welcome &#x1f609;.
 
Please [create an issue](https://github.com/Apicurio/apicurio-registry-operator/issues/new) on GitHub if you come across any problems.

You can find more help and information about the project in the [Troubleshooting](docs/Troubleshooting)  
and [About](docs/About) pages.

*Note: To view older versions of these pages, checkout the appropriate tag.*
