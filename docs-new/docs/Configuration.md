---
layout: default
title: Configuration
nav_order: 3
has_children: true
---

# Configuration
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Persistence Options

The main decision to do when deploying Apicurio Registry is which persistence backend to use. 
Following options are available: 

 - In-memory - data are stored in RAM, replicated on each Apicurio Registry node. This is the easiest deployment to setup, 
 but we don't recommend to use it in production.
 - JPA (PostgreSQL) - data are stored in a relational database
 - Kafka (basic) - data are stored using plain Kafka
 - Kafka Steams - data are stored using Kafka Streams. This option is suggested for most production deployments.
 - Infinispan - data are stored in a similar way as in the in-memory option, but using embedded Infinispan.

Most of these options (JPA, Kafka, Streams) require the storage is set up as a prerequisite. This will change as the Operator improves.
You will find more information in the given sub-pages. 

## Apicurio Registry Custom Resource

The Operator represents a deployment of Apicurio Registry using a single custom resource (CR).

This is the full tree of the configuration options present in the `ApicurioRegistry` custom resource definition (CRD). 
Short description of each one follows.

*Note: These configuration options may change or be renamed until the operators leaves the alpha development stage.*

### Spec

`Spec` is the part of CR that users edit to provide desired state for the Operator to achieve.

```yaml
spec:
  configuration:
    persistence: <string>
    dataSource:
      url: <string>
      userName: <string>
      password: <string>
    kafka:
      bootstrapServers: <string>
    streams:
      bootstrapServers: <string>
      applicationId: <string>
      applicationServerPort: <string>
      security:
        tls:
          truststoreSecretName: <string>
          keystoreSecretName: <string>
        scram:
          mechanism: <string>
          truststoreSecretName: <string>
          user: <string>
          passwordSecretName: <string>
    infinispan:
      clusterName: <string>
    ui:
      readOnly: <string>
    logLevel: <string>
  deployment:
    replicas: <int32>
    host: <string>
```

| Configuration Option | Type | Default Value | Description
|---|---|---|---|
|configuration|-|-|Section for configuration of an Apicurio Registry application
|configuration/persistence|string|mem|Persistence backend to use; one of: mem, jpa, kafka, streams, infinispan|
|configuration/**dataSource**|-|-|Database connection configuration for JPA persistence backend|
|configuration/**dataSource**/url|string|*required*|Database connection URL string|
|configuration/**dataSource**/userName|string|*required*|Database connection user|
|configuration/**dataSource**/password|string|*empty*|Database connection password|
|configuration/**kafka**|-|-|Kafka backend configuration section|
|configuration/**kafka**/bootstrapServers|string|*required*|Kafka bootstrap server URL|
|configuration/**streams**|-|-|Kafka Streams persistence backend configuration|
|configuration/**streams**/bootstrapServers|string|*required*|Kafka bootstrap server URL, for Streams persistence backend|
|configuration/**streams**/applicationId|string|*ApicurioRegistry CR name*|Kafka Streams application ID|
|configuration/**streams**/applicationServerPort|string|9000|-|
|configuration/**streams**/security/**tls**|-|-|Section to enable and configure TLS authentication for Kafka Streams persistence backend|
|configuration/**streams**/security/**tls**/truststoreSecretName|string|*required*|Name of a secret containing TLS truststore for Kafka|
|configuration/**streams**/security/**tls**/keystoreSecretName|string|*required*|Name of a secret containing user's TLS keystore|
|configuration/**streams**/security/**scram**/truststoreSecretName|string|*required*|Name of a secret containing TLS truststore for Kafka|
|configuration/**streams**/security/**scram**/user|string|*required*|SCRAM user name|
|configuration/**streams**/security/**scram**/passwordSecretName|string|*required*|Name of a secret containing SCRAM user password|
|configuration/**streams**/security/**scram**/mechanism|string|SCRAM-SHA-512|SASL mechanism|
|configuration/**infinispan**|-|-|Infinispan persistence configuration section|
|configuration/**infinispan**/clusterName|string|*ApicurioRegistry CR name*|Infinispan cluster name|
|configuration/**ui**|-|-|Web UI settings|
|configuration/**ui**/readOnly|string|false|Set web UI to read-only mode|
|configuration/logLevel|string|INFO|Operand log level; one of: INFO, DEBUG|
|deployment|-|-|Section for operand deployment settings|
|deployment/**replicas**|positive integer|1|Number of Apicurio Registry pods to deploy|
|deployment/**host**|string|*auto-generated (from ApicurioRegistry CR name and namespace)*|Host/URL where the Apicurio Registry UI and API is available|

*Note: If an option is marked as **required**, it may be conditional on other configuration options enabled. Empty value may be syntactically accepted, but the operator will not perform given action.* 

### Status

This is the section of the CR, managed by the Operator, that contains description of the current state.

```yaml
status:
  image: <string>
  deploymentName: <string>
  serviceName: <string>
  ingressName: <string>
  replicaCount: <int32>
  host: <string>
```

| Status Entry | Type | Description
|---|---|---|
|image|string|Operand image that the operator deploys. May change based on the persistence option selected in the configuration.|
|deploymentName|string|Name of the `Deployment` or `DeploymentConfig` managed by the Operator, used to deploy the Registry. |
|serviceName|string|Name of the `Service` managed by the Operator, to expose the Operand as a service.|
|ingressName|string|Name of the `Ingress` managed by the Operator, to make the Registry accessible via HTTP. On OCP, a `Route` is created as well.|
|replicaCount|int32|Number of the Operand pods deployed.|
|host|string|URL where the Registry is accessible.|

### Labels

All managed resources are usually labeled by the following:

| Label | Description
|---|---|
|app|Name of the Apicurio Registry deployment the resource belongs to, based on the name of the given `ApicurioRegsitry` CR.| 

## Managed Resources

Resources managed by the operator:

- Deployment (Kubernetes) or DeploymentConfig (OpenShift)
- Service
- Ingress, and a Route (OpenShift)
- PodDisruptionBudget
