---
layout: default
title: Kafka Streams
parent: Configuration
---

# Kafka Streams Persistence Configuration
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}


## Prerequisites

Install the *Apicurio Registry Operator* from Operator Hub or manually using `kubectl apply`.

## Kafka setup

Install the *Strimzi* operator from the Operator Hub.

---

Notice:

We are going to be using *Strimzi* operator to provide us with a Kafka cluster.

If you have a cluster deployed in another way, you can still use it with the Apicurio Registry Operator,
but you would have to provide the required resources, such as Kafka topics or auth. secrets manually.

## Plain (no auth)

### Kafka Cluster

We can provision a Kafka cluster by creating a `Kafka` resource. You can do that in the UI by navigating to *Installed Operators*, then selecting *Strimzi* operator details, and opening the *Kafka* tab. Click on the *Create Kafka* button. You can use the default value, which should look something like:

```
apiVersion: kafka.strimzi.io/v1beta1
kind: Kafka
metadata:
  name: my-cluster
  namespace: registry-example-streams-plain
spec:
  kafka:
    version: 2.5.0
    replicas: 3
    listeners:
      plain: {}
      tls: {}
    config:
      offsets.topic.replication.factor: 3
      transaction.state.log.replication.factor: 3
      transaction.state.log.min.isr: 2
      log.message.format.version: '2.5'
    storage:
      type: ephemeral
  zookeeper:
    replicas: 3
    storage:
      type: ephemeral
  entityOperator:
    topicOperator: {}
    userOperator: {}
```

*Note: Your namespace may be different*

After the cluster is ready, you can open the *Kafka* resource and examine the `status` block.
There should be a `bootstrapServers` property, which we will later use to deploy *Apicurio Registry*. Example:

```
status:
  conditions:
    ...
  listeners:
    - addresses:
        - host: my-cluster-kafka-bootstrap.registry-example-streams-plain.svc
          port: 9092
      bootstrapServers: 'my-cluster-kafka-bootstrap.registry-example-streams-plain.svc:9092'
      type: plain
  ...
```

### Kafka Topics

Before we can do a basic deployment, two topics have to be created: `global-id-topic` and `storage-topic`. These names are not configurable at the moment.

As before, navigate to *Kafka Topic* tab, and click on *Create Kafka Topic*:

```
apiVersion: kafka.strimzi.io/v1beta1
kind: KafkaTopic
metadata:
  name: global-id-topic
  labels:
    strimzi.io/cluster: my-cluster
  namespace: registry-example-streams-plain
spec:
  partitions: 2
  replicas: 1
  config:
    retention.ms: 604800000
    segment.bytes: 1073741824
```

Change the topic name to `global-id-topic`, and optionally decrease partition and replica count to minimum, which is going to be sufficient for our example deployment.

Do the same to create the other topic.

### Deploy the Apicurio Registry

Switch to the *Apicurio Registry Operator*, and in the *ApicurioRegistry* tab, click on *Create ApicurioRegistry*.
Use the following spec, but use your value for the `bootstrapServers` property. (See also the example yaml files.)

```
apiVersion: apicur.io/v1alpha1
kind: ApicurioRegistry
metadata:
  name: example-apicurioregistry
spec:
  configuration:
    persistence: "streams"
    streams:
      bootstrapServers: "my-cluster-kafka-bootstrap.registry-example-streams-plain.svc:9092"
```

After a few minutes, you should see a *Route* being created, where you can access the application.

## TLS Security

You can configure *Strimzi* and *Apicurio Registry Operator* to use an encrypted TLS connection.

### Kafka Cluster

As before, you have to deploy a new cluster, but in this case, you need to configure that TLS authenticatione is used:

```
apiVersion: kafka.strimzi.io/v1beta1
kind: Kafka
metadata:
  name: my-cluster
  namespace: registry-example-streams-tls
spec:
  kafka:
    authorization:
      type: simple
    version: 2.5.0
    replicas: 3
    listeners:
      plain: {}
      tls:
        authentication:
          type: tls
    config:
      offsets.topic.replication.factor: 3
      transaction.state.log.replication.factor: 3
      transaction.state.log.min.isr: 2
      log.message.format.version: '2.5'
    storage:
      type: ephemeral
  zookeeper:
    replicas: 3
    storage:
      type: ephemeral
  entityOperator:
    topicOperator: {}
    userOperator: {}
```

See the `authorization` and `tls` sections.

### Auth setup

Configure the topics as before, but in addition, we need to create a *Kafka User* resource, to configure authentication and authorization for our Apicurio Registry user. This is an example `spec` block:

```
spec:
  authentication:
    type: tls
  authorization:
    acls:
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: topic
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: cluster
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: transactionalId
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: group
    type: simple
```

You can specify a user name in the `metadata` section, we will use the default `my-user`.

*Note: You can (should) configure the authorization specifically for the topics and resources that Apicurio Registry needs, but this is a more compact example version.*

Afterwards, *Strimzi* will create two secrets we are going to need to enable Apicurio Registry to connect.

Navigate to *Workloads* then *Secrets*, where you should find two secrets using a filter: `my-cluster-cluster-ca-cert` containing PKCS12 truststore for the Kafka cluster, and `my-user`, that contains user's keystore. Secret names can vary if your cluster or user is named differently.

---

Note: If you are creating the secrets manually, they must contain following key-value pairs:

- my-cluster-ca-cert
  - `ca.p12` - the truststore in PKCS12 format
  - `ca.password` - truststore password
- my-user
  - `user.p12` - the keystore in PKCS12 format
  - `user.password` - keystore password


### Deploy the Apicurio Registry

Use the following example configuration to deploy the registry. Important, you have to use a different `bootstrapServers` address,
one that supports TLS. You can find it in the *Kafka* resource as before, but it has `type: tls`.

```
apiVersion: apicur.io/v1alpha1
kind: ApicurioRegistry
metadata:
  name: example-apicurioregistry
spec:
  configuration:
    persistence: "streams"
    streams:
      bootstrapServers: "my-cluster-kafka-bootstrap.registry-example-streams-tls.svc:9093"
      security:
        tls:
          keystoreSecretName: my-user
          truststoreSecretName: my-cluster-cluster-ca-cert
```

## TLS+SCRAM Security

### Kafka Cluster

As before, you have to deploy a new cluster, but in this case, you need to configure that SCRAM-SHA-512 authentication is used:

```
apiVersion: kafka.strimzi.io/v1beta1
kind: Kafka
metadata:
  name: my-cluster
  namespace: registry-example-streams-tls
spec:
  kafka:
    authorization:
      type: simple
    version: 2.5.0
    replicas: 3
    listeners:
      plain: {}
      tls:
        authentication:
          type: scram-sha-512
    config:
      offsets.topic.replication.factor: 3
      transaction.state.log.replication.factor: 3
      transaction.state.log.min.isr: 2
      log.message.format.version: '2.5'
    storage:
      type: ephemeral
  zookeeper:
    replicas: 3
    storage:
      type: ephemeral
  entityOperator:
    topicOperator: {}
    userOperator: {}
```

See the `authorization` and `tls` sections.

### Auth setup

Configure the topics as before, but in addition, we need to create a *Kafka User* resource, to configure authentication and authorization for our Apicurio Registry user. This is an example `spec` block:

```
spec:
  authentication:
    type: scram-sha-512
  authorization:
    acls:
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: topic
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: cluster
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: transactionalId
      - operation: All
        resource:
          name: '*'
          patternType: literal
          type: group
    type: simple
```

See the `authentication` part.

As before, *Strimzi* will create two secrets we are going to need to enable Apicurio Registry to connect.

Navigate to *Workloads* then *Secrets*, where you should find two secrets using a filter: `my-cluster-cluster-ca-cert` containing PKCS12 truststore for the Kafka cluster, and `my-user`, that contains user's SCRAM password. Secret names can vary if your cluster or user is named differently.

---

Note: If you are creating the secrets manually, they must contain following key-value pairs:

- my-cluster-ca-cert
  - `ca.p12` - the truststore in PKCS12 format
  - `ca.password` - truststore password
- my-user
  - `password` - user's password

### Deploy the Apicurio Registry

Use the following example configuration to deploy the registry. Important, you have to use a different `bootstrapServers` address,
one that supports TLS. You can find it in the *Kafka* resource as before, but it has `type: tls`.

```
apiVersion: apicur.io/v1alpha1
kind: ApicurioRegistry
metadata:
  name: example-apicurioregistry
spec:
  configuration:
    persistence: "streams"
    streams:
      bootstrapServers: "my-cluster-kafka-bootstrap.registry-example-streams-scram.svc:9093"
      security:
        scram:
          truststoreSecretName: my-cluster-cluster-ca-cert
          user: my-user
          passwordSecretName: my-user
```
