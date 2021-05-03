# Apicurio Registry Installation and Example Files

## How to Install the Operator

Create a namespace for the installation, e.g., `apicurio-registry-operator-namespace`:

```
NAMESPACE="apicurio-registry-operator-namespace"
kubectl create namespace "$NAMESPACE"
```

All Kubernetes resources required for the installation are present inside `default-install.yaml`.

If you use a different namespace, find and replace `apicurio-registry-operator-namespace`
in `default-install.yaml` with the namespace you want to use.

Apply the installation file:

```
kubectl apply -f default-install.yaml
```

or apply the installation file for a different namespace:

```
cat default-install.yaml | sed "s/apicurio-registry-operator-namespace/$NAMESPACE/g" | kubectl apply -f -
```

## How to Install the Registry

The registry supports the following persistence options:

* In-Memory (`mem`)
* PostgreSQL (`sql`)
* Kafka (`kafkasql`)

Examples of ApicurioRegistry Custom Resources configured for different persistence solutions can be found in the `examples/` folder. Apply one of them:

```
kubectl apply -f ./examples/apicurioregistry_<PERSISTENCE>_cr.yaml -n "$NAMESPACE"
```
