# Installation

You can install Apicurio Registry Operator using:

* [OperatorHub](https://operatorhub.io/operator/apicurio-registry) (also available on OpenShift)
* Install file from the distribution archive that is attached to the corresponding GitHub release. Follow the instructions in the included README file.
* Install file from the [Apicurio Registry Operator repository](https://github.com/Apicurio/apicurio-registry-operator/tree/main/install).

# Features

* Deployed Apicurio Registry version is `2.4.12.Final`.

* Support Kubernetes `1.25` by working with both `v1beta1` and `v1` versions of *PodDisruptionBudget*.

* Support configuration of custom environment variables for Apicurio Registry using `spec.configuration.env` field in the *ApicurioRegistry* CR.

* You can provide a secret containing a TLS certificate and key using `spec.configuration.security.https.secretName` field, and Apicurio Registry Operator will configure Apicurio Registry to support HTTPS connections on port `8443`.

* You can set log level for non-library Apicurio Registry components separately using `spec.configuration.registryLogLevel` field.

* You can set log level for the Apicurio Registry Operator by configuring an environment variable `LOG_LEVEL` in its *Deployment*. Supported values are `debug`, `info` (default), `warn`, and `error`.

* Support configuring custom affinity, tolerations, and image pull secrets, using the `spec.deployment.affinity`, `spec.deployment.tolerations`, and `spec.deployment.imagePullSecrets` fields, respectively.

* You can configure custom Apicurio Registry image using the `spec.deployment.image` field.

* Support configuring custom annotations and labels for the Apicurio Registry pod using `spec.deployment.metadata.annotations` and `spec.deployment.metadata.labels` fields.

* You can prevent Apicurio Registry Operator from managing the following resources:
    * *Ingress*
    * *NetworkPolicy*
    * *PodDisruptionBudget*

  in case you need to manage them yourself, using `spec.deployment.managedResources.*` fields.

* The *ApicurioRegistry* CRD now contains the `spec.deployment.podTemplateSpecPreview` field, which has the same structure as the field `spec.template` in a Kubernetes *Deployment* resource. With some restrictions, you can use this field to customize the Apicurio Registry *Deployment*. This is a Technology Preview feature.

* The Apicurio Registry Operator now sets the `CORS_ALLOWED_ORIGINS` environment variable, based on the value of the `spec.deployment.host` field. This environment variable controls the `Access-Control-Allow-Origin` header sent by Apicurio Registry. You can override the default value using the `spec.configuration.env` field.

This list may not be exhaustive. For more details, review the documentation included in the distribution archive or at [Apicurio Registry documentation page](https://www.apicur.io/registry/docs/) (select the appropriate Apicurio Registry Operator version in the lower left corner).
