# Installation

You can install Apicurio Registry Operator using:

* [OperatorHub](https://operatorhub.io/operator/apicurio-registry) (also available on OpenShift)
* Install file from the distribution archive that is attached to the corresponding GitHub release. Follow the instructions in the included README file.
* Install file from the [Apicurio Registry Operator repository](https://github.com/Apicurio/apicurio-registry-operator/tree/main/install).

# Features

* Deployed Apicurio Registry version is `2.5.9.Final`.

* The `JAVA_OPTIONS` environment variable no longer works as expected. You can use the `JAVA_OPTS_APPEND` environment variable instead. The `JAVA_OPTS` environment variable is also available, which replaces the default content of Java options. However, it is best to append using the `JAVA_OPTS_APPEND` environment variable if possible. Apicurio Registry Operator has been updated to use the new environment variables.

* Apicurio Registry Operator supports configuration of the SQL data source using environment variables, as an alternative to the `spec.configuration.sql.dataSource` fields in the `ApicurioRegistry` custom resource. This allows users to provide SQL credentials using Kubernetes secrets instead of in plaintext.

  Apicurio Registry Operator has been improved in this version to better support this use case. Users are now able to use both the `spec.configuration.sql.dataSource` and `spec.configuration.env` fields to define parts of the configuration. For example, this  is now valid:

  ```yaml
  apiVersion: registry.apicur.io/v1
  kind: ApicurioRegistry
  metadata:
    name: myregistry
  spec:
    configuration:
    persistence: sql
    sql:
      dataSource:
        url: "jdbc:postgresql://..."
        userName: "postgres-user"  
    env:
      - name: REGISTRY_DATASOURCE_PASSWORD
        valueFrom:
        secretKeyRef:
          name: postgres-secret
          key: password
  ```

  Apicurio Registry Operator detects this type of configuration and applies it immediately without a need for additional user intervention.

* Updated version of `go` to `1.20`.

* Updated dependency versions.

This list may not be exhaustive. For more details, review the documentation included in the distribution archive or at [Apicurio Registry documentation page](https://www.apicur.io/registry/docs/) (select the appropriate Apicurio Registry Operator version in the lower left corner).
