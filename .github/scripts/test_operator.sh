#!/bin/bash
set -e -a

OPERATOR_METADATA_IMAGE=docker.io/apicurio/apicurio-registry-operator-metadata:latest-dev
BUNDLE_URL=${PWD}/docs/resources/install-dev.yaml

git clone https://github.com/Apicurio/apicurio-registry-k8s-tests-e2e.git

pushd apicurio-registry-k8s-tests-e2e

./scripts/install_kind.sh

make run-operator-ci

popd

set +e +a