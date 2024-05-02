# Installation

You can install Apicurio Registry Operator using:

* [OperatorHub](https://operatorhub.io/operator/apicurio-registry) (also available on OpenShift)
* Install file from the distribution archive that is attached to the corresponding GitHub release. Follow the instructions in the included README file.
* Install file from the [Apicurio Registry Operator repository](https://github.com/Apicurio/apicurio-registry-operator/tree/main/install).

# Features

* Deployed Apicurio Registry version is `2.5.11.Final`.

* Handle the `JAVA_OPTIONS` variable name change in a backwards compatible way. Requiring the user to change the variable name causes issues when automatic OLM updates are enabled. This update ensures that the operator translates use of `JAVA_OPTIONS` to `JAVA_OPTS_APPEND` automatically. Users are still encouraged to rename the variable.

* Update descriptions in the CSV, add new annotations required by Kubernetes and Openshift. 

This list may not be exhaustive. For more details, review the documentation included in the distribution archive or at [Apicurio Registry documentation page](https://www.apicur.io/registry/docs/) (select the appropriate Apicurio Registry Operator version in the lower left corner).
