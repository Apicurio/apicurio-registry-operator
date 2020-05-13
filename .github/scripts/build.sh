#!/bin/bash
set -e

VERSION=$(sed -n 's/^.*Version.*=.*"\([0-9\.]*\)".*$/\1/p' ./version/version.go)
if [ -z "$VERSION" ]
then
    echo "Could not read project version."
    exit 1
fi

OPERATOR_IMAGE_NAME="${IMAGE_REGISTRY}/${IMAGE_REGISTRY_ORG}/apicurio-registry-operator"
OPERATOR_IMAGE="${OPERATOR_IMAGE_NAME}:${VERSION}"

operator-sdk generate k8s
operator-sdk generate openapi
operator-sdk build "${OPERATOR_IMAGE}"
docker tag "${OPERATOR_IMAGE}" "${OPERATOR_IMAGE_NAME}:${TAG}"

if [[ -v PUSH ]]
then
    echo "Logging in to image registry and pushing image"
    docker login -u "${REGISTRY_USER}" -p "${REGISTRY_PASS}" "${IMAGE_REGISTRY}"

    docker push "${OPERATOR_IMAGE}"
    docker push "${OPERATOR_IMAGE_NAME}:${TAG}"
fi

