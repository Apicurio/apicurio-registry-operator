#!/bin/bash
set -e -a

VERSION=$(sed -n 's/^.*Version.*=.*"\(.*\)".*$/\1/p' ./version/version.go)

echo $VERSION

OPERATOR_IMAGE="quay.io/apicurio/apicurio-registry-operator:$VERSION"
OPERATOR_METADATA_IMAGE="quay.io/apicurio/apicurio-registry-operator-bundle:$VERSION"

BUNDLE_URL=${PWD}/dist/install.yaml
OPERATOR_PROJECT_DIR=${PWD}

make dist

git clone https://github.com/Apicurio/apicurio-registry-k8s-tests-e2e.git

pushd apicurio-registry-k8s-tests-e2e

./scripts/install_kind.sh

make run-operator-ci

popd

set +e +a
