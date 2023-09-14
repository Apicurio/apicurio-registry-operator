#!/bin/bash
set -e -a

if [[ ! $OPERATOR_VERSION ]]; then
  OPERATOR_VERSION=$(sed -n 's/.*Version.*=.*"\(.*\)".*/\1/p' ./version/version.go)
fi
OPERATOR_VERSION_SUFFIX=$(echo "$OPERATOR_VERSION" | sed -n 's/^[0-9\.]*-\([^-+]*\).*$/-\1/p')

if [[ ! $OPERAND_VERSION ]]; then
  OPERAND_VERSION="2.x"
fi
LC_OPERAND_VERSION=$(echo $OPERAND_VERSION | tr A-Z a-z)

if [[ ! $OPERATOR_IMAGE_REPOSITORY ]]; then
  OPERATOR_IMAGE_REPOSITORY="quay.io/apicurio"
fi

PACKAGE_VERSION="$OPERATOR_VERSION-v$LC_OPERAND_VERSION"

OPERATOR_IMAGE="$OPERATOR_IMAGE_REPOSITORY/apicurio-registry-operator:$OPERATOR_VERSION"
BUNDLE_IMAGE="$OPERATOR_IMAGE_REPOSITORY/apicurio-registry-operator-bundle:$PACKAGE_VERSION"
CATALOG_IMAGE="$OPERATOR_IMAGE_REPOSITORY/apicurio-registry-operator-catalog:latest$OPERATOR_VERSION_SUFFIX"

OPERATOR_PROJECT_DIR=$(pwd)

make dist

git clone https://github.com/Apicurio/apicurio-registry-k8s-tests-e2e.git
pushd apicurio-registry-k8s-tests-e2e

git checkout master
./scripts/install_kind.sh
make run-operator-ci

popd

git reset --hard
git clean -df
rm -rf apicurio-registry-k8s-tests-e2e dist

set +e +a
