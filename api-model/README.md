# Operator CRD Model

## Build

In the repository root, run:

```shell
make bundle
export PACKAGE_VERSION=$(make get-variable-package-version)
```

then `cd api-model` and run:

```shell
mvn clean install -DpackageVersion=$PACKAGE_VERSION
```
