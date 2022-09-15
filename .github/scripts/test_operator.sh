#!/bin/bash
set -e -a

VERSION=$(sed -n 's/^.*Version.*=.*"\(.*\)".*$/\1/p' ./version/version.go)
DASH_VERSION = $(shell echo "$(VERSION)" | sed -n 's/^[0-9\.]*-\([^-+]*\).*$$/-\1/p')

echo $VERSION
echo $DASH_VERSION

OPERATOR_IMAGE="quay.io/apicurio/apicurio-registry-operator:$VERSION"
OPERATOR_METADATA_IMAGE="quay.io/apicurio/apicurio-registry-operator-bundle:$VERSION"
OLM_CSV=apicurio-registry-operator.v1.1.0-dev-v2.x
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
