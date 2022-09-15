#!/bin/bash
set -e -a

VERSION=$(sed -n 's/^.*Version.*=.*"\(.*\)".*$/\1/p' ./version/version.go)
DASH_VERSION=$(echo "$VERSION" | sed -n 's/^[0-9\.]*-\([^-+]*\).*$/-\1/p')

echo $VERSION
echo $DASH_VERSION

OPERATOR_IMAGE="quay.io/apicurio/apicurio-registry-operator:$VERSION"
CSV_VERSION=1.1.0-dev-v2.x
OPERATOR_METADATA_IMAGE="quay.io/apicurio/apicurio-registry-operator-bundle:$CSV_VERSION"
OLM_CSV="apicurio-registry-operator.v$CSV_VERSION"
CATALOG_SOURCE_IMAGE="quay.io/apicurio/apicurio-registry-operator-catalog:latest$DASH_VERSION"

BUNDLE_URL=${PWD}/dist/install.yaml
OPERATOR_PROJECT_DIR=${PWD}

make dist

git clone https://github.com/Apicurio/apicurio-registry-k8s-tests-e2e.git

pushd apicurio-registry-k8s-tests-e2e

./scripts/install_kind.sh

make run-operator-ci

popd

set +e +a
