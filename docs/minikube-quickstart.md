Minikube Quickstart
===

Requirements
---
* Make sure you are able to build the Apicutio Registry ("AR") operator. See the project README doc for details.
* Install a [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) cluster.

1\. Minikube
---
Start MK with extra memory:

`$ minikube start --memory=6144`

2\. Ingress
---

Make sure ingress addon is enabled:

`$ minikube addons enable ingress`

Get the cluster IP:

```
$ minikube ip
192.168.99.111
```

Edit `/etc/hosts` to resolve registry host to the cluster IP:

```
$ cat /etc/hosts
[...]
192.168.99.111   registry.example.com
```

Do not forget to use the same host in the registry CR.

3\. Build
---

Choose a registry to store your build of the operator, e.g. [quay.io](quay.io) and build the operator:

```
$ ./build.sh build -r "$REGISTRY"
$ ./build.sh push -r "$REGISTRY"
```

See the project README doc for details.

4\. In-memory Deployment
---

To deploy AR with the in-memory storage, you don't need to deploy anything else, but you have to use 1 replica deployment or lose data consistency.
You can use one of the example CRs:

```
$ ./build.sh mkdeploy -r "$REGISTRY" --cr ./docs/example-cr/in-memory.yaml
```

5\. Test Queries
---

Try the following HTTP requests to test the deployment:

```
curl -v http://registry.example.com/health
curl -v http://registry.example.com/health/ready
curl -v http://registry.example.com/health/live
```

Create a test artifact:

```
curl -X POST -H "Content-Type: application/json" \
     -H "X-Registry-ArtifactType: JSON" \
     -H "X-Registry-ArtifactId: test1" \
     -d '{"type": "cat", "color": "black"}' \
     http://registry.example.com/artifacts
```

6\. Streams Deployment
---

Use Strimzi operator to deploy Kafka services. Follow this [guide](https://strimzi.io/quickstarts/minikube/).

Make sure Kafka has been deployed:

```
$ kubectl -n kafka get pods
NAME                                          [...]
my-cluster-entity-operator-5d54fdbd94-9qmzb
my-cluster-kafka-0
my-cluster-zookeeper-0
strimzi-cluster-operator-6975d8874f-hhv4m
```

Create two topics, `global-id-topic` and `storage-topic` required by the application:

```
$ cat topic1.yaml
apiVersion: kafka.strimzi.io/v1beta1
kind: KafkaTopic
spec:
  partitions: 2
  replicas: 1
  topicName: global-id-topic
$ kubectl create -f topic1.yaml
$ kubectl create -f topic2.yaml
```

If you have an existing AR deployment, you can either edit the CR,
or run `./build.sh mkundeploy -r "$REGISTRY"` to remove it.

Verify that the `bootstrapServers` configuration option in  the CR is correct.
Run `minikube service -n kafka list`, the URL should be `<name>.<namespace>.svc:9092`.

```
$ ./build.sh mkdeploy -r "$REGISTRY" --cr ./docs/example-cr/streams.yaml
```

7\. Kafka Deployment
---

The steps are very similar to the Streams deployment,
but be careful not to reuse the same topics (without deleting and recreating).

8\. JPA Deployment
---

Deploy the [Postgresql Operator](https://github.com/zalando/postgres-operator/blob/master/docs/quickstart.md).
Create the `registry` database before using the example CR.

```
$ kubectl port-forward acid-minimal-cluster-0 6543:5432 &
$ export PGPASSWORD=$(...)
$ psql -U postgres -h 127.0.0.1 -p 6543
postgres=# create database registry;
```

