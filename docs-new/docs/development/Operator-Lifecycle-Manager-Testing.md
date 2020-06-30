---
layout: default
title: Operator Lifecycle Manager Testing
parent: Development
---

# Operator Lifecycle Manager Testing
{: .no_toc }

The final released Apicurio Registry will be installed by users from the Operator Hub.
You can test unreleased operators by adding them manually to the Operator Hub (OLM Catalog) 
in you own OpenShift (Kubernetes) instance.

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

*Note: The following guide is written for OpenShift, but is applicable for Kubernetes as well.
In case of Kubernetes, you have to make sure the OLM operator is installed first.*

## Setting up the app registry on quay.io

1. Register on quay.io

1. Retrieve your `QUAY_TOKEN` using this command (enter your credentials):

   ```
   echo -n 'Your quay.io username: ' \
     && read QUAY_USERNAME \
     && echo -n 'Your quay.io password: ' \
     && export QUAY_TOKEN=$(curl --silent -H "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '{"user":{"username":"'"${QUAY_USERNAME}"'","password":"'"$(read -s PW && echo -n $PW)"'"}}' | sed -E 's/.*\"(basic .*)\".*/\1/')
   ```

1. Verify by `echo $QUAY_TOKEN`.

## Prepare images

1. Push the images you want to test to your quay.io repository. 
   There is 1 image for the operator, and several operand images (Streams, JPA, Infinispan, ...).

1. Prepare your manifests files for installation to quay.io. 
   The `deploy/olm-catalog` directory contains the latest manifest files that have been published in the Operator Hub.
   You can create a new CSV version for testing.  

1. Update image references to those that have been pushed to your quay.io namespace.

   ```
   # Replace operator image!
   image: 'quay.io/<namespace>/apicurio-registry-operator:0.0.4-dev'
   ```
    
   *Note: We will use tags here, but image hashes are used in releases.*

1. Do the same for operand images in the `env:` and related images section. Example:
    
   ```
   env:
     # Replace these!
     - name:  REGISTRY_IMAGE_STREAMS
       value: "quay.io/<namespace>/apicurio-registry-streams:1.2.3.Final"
     [...]
   ```

## Upload data to quay.io

1. Install the `operator-courier` tool, using Python 3 installer `pip3 install operator-courier`.

1. Set up the configuration variables, like in the example (replace with your own values):

   ```
   export OPERATOR_DIR=`pwd`/deploy/olm-catalog/apicurio-registry
   export QUAY_NAMESPACE="<namespace>"
   export PACKAGE_NAME="apicurio-registry-operator"
   export PACKAGE_VERSION="0.0.4-dev"
   export TOKEN="$QUAY_TOKEN"
   ```

1. Run operator courier: `operator-courier --verbose  push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"`

1. Make sure that the manifests have been uploaded by visiting `https://quay.io/application/`.

1. You can mark the resources at quay.io as public to avoid setting up auth.

## Configuring OpenShift

1. Use the `operator-source.yaml` file, and apply it to your cluster using `oc -n openshift-marketplace apply -f operator-source.yaml`.

   Example:
    
   ```yaml
   apiVersion: operators.coreos.com/v1
   kind: OperatorSource
   metadata:
     name: test-operators
     namespace: openshift-marketplace
   spec:
     type: appregistry
     endpoint: https://quay.io/cnr
     registryNamespace: <namespace>
   ```

1. You should see your version of the `Apicurio Registry Operator` in the Operator Hub on your cluster within a few minutes.
