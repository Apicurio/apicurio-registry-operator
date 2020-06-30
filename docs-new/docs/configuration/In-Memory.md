---
layout: default
title: In-Memory
parent: Configuration
---

# In-Memory Persistence Configuration
{: .no_toc }
<!--
## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Overview
-->


In-Memory is the simplest persistence option to use. 
The Operator will deploy Registry in this way if you don't provide any configuration, so the CR is very simple:

```yaml
apiVersion: apicur.io/v1alpha1
kind: ApicurioRegistry
metadata:
  name: example-apicurioregistry
# spec:
# No config required
```  

You can deploy this example using:

```
export GIT_REF="master" &&
curl -sSL "https://raw.githubusercontent.com/apicurio/apicurio-registry-operator/$GIT_REF/docs/resources/example-cr/in-memory.yaml" | kubectl apply -f -
``` 

The major downside of this option is that the data are not shared across replicas,
so using multiple replicas is not recommended. 
