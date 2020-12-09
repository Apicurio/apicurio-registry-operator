#!/bin/bash
set -e -a

VERSION=$(sed -n 's/^.*Version.*=.*"\(.*\)".*$/\1/p' ./version/version.go)

echo $VERSION

OPERATOR_IMAGE="${IMAGE_REGISTRY}/${IMAGE_REGISTRY_ORG}/apicurio-registry-operator:$VERSION"
OPERATOR_METADATA_IMAGE="${IMAGE_REGISTRY}/${IMAGE_REGISTRY_ORG}/apicurio-registry-operator-metadata:$VERSION"

BUNDLE_URL=${PWD}/docs/resources/install-dev.yaml

git clone https://github.com/Apicurio/apicurio-registry-k8s-tests-e2e.git

pushd apicurio-registry-k8s-tests-e2e

./scripts/install_kind.sh

make run-operator-ci

popd

set +e +a