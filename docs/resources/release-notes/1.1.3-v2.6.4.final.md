# Installation

You can install Apicurio Registry Operator using:

* [OperatorHub](https://operatorhub.io/operator/apicurio-registry) (also available on OpenShift)
* Install file from the distribution archive that is attached to the corresponding GitHub release. Follow the instructions in the included README file.
* Install file from the [Apicurio Registry Operator repository](https://github.com/Apicurio/apicurio-registry-operator/tree/main/install).

# Features

* Deployed Apicurio Registry version is `2.6.4.Final`.

* Fix: PTS (podTemplateSpec) change not detected when Deployment is deleted or edited.

* Fix: Initialization check failing due to cyclic redirect.

* Fix: Env. variables not applied when Deployment is deleted. Improve detection of when Deployment does not contain correct env. variables. Remove deprecated feature that allowed users to manually edit Deployment to specify env. variables - this has caused the code to be more complex and increased chance of bugs, and is not feasible to support it together with the PTS feature.

* Improve panic messages

* Fix: Update Quarkus HTTP TLS options. With a Quarkus version upgrade, the supporting TLS options have changed. It no longer uses `quarkus.http.ssl.certificate.key-file` and `quarkus.http.ssl.certificate.file`, but `*.files`. See https://quarkus.io/guides/http-reference#configuring-the-http-server-directly .

This list may not be exhaustive. For more details, review the documentation included in the distribution archive or at [Apicurio Registry documentation page](https://www.apicur.io/registry/docs/) (select the appropriate Apicurio Registry Operator version in the lower left corner).
