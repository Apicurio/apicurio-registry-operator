---
layout: default
title: Infinispan
parent: Configuration
---

# Infinispan Persistence Configuration
{: .no_toc }

Infinispan persistence option is an improvement upon the in-memory storage, 
because the data is replicated between all Registry replicas.
This is achieved using embedded Infinispan in-memory data grid.
The configuration stays simple, although that may change if we transition 
to use an external Infinispan cluster in the future.

Following is the example CR:

```yaml
apiVersion: apicur.io/v1alpha1
kind: ApicurioRegistry
metadata:
  name: example-apicurioregistry
spec:
  configuration:
    persistence: "infinispan"
    infinispan: # Currently, registry uses an embedded version of Infinispan
      clusterName: "example-apicurioregistry"
      # ^ Optional
```

And you can deploy this example using:

```
export GIT_REF="master" &&
curl -sSL "https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/$GIT_REF/docs/resources/example-cr/infinispan.yaml" | kubectl apply -f -
```

