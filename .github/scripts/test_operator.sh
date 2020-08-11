#!/bin/bash
set -e

git clone git@github.com:Apicurio/apicurio-registry-k8s-tests-e2e.git

pushd apicurio-registry-k8s-tests-e2e

./scripts/install_kind.sh

make run-operator-ci

popd