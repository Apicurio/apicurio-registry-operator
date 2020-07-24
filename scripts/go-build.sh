#!/bin/sh

source ./scripts/go-mod-env.sh

  echo  OPERATOR_IMAGE_REPOSITORY = $OPERATOR_IMAGE_REPOSITORY



IMAGE=apicurio-registry-operator
TAG=$(sed -n 's/^.*Version.*=.*"\(.*\)".*$/\1/p' ./version/version.go)


CFLAGS="--redhat --build-tech-preview"

./scripts/go-gen.sh

if [[ -z ${CI} ]]; then
    ./scripts/go-test.sh
    operator-sdk build ${OPERATOR_IMAGE_REPOSITORY}/${IMAGE}:${TAG}
   
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o build/_output/bin/apicurio-registry-operator github.com/Apicurio/apicurio-registry-operator/cmd/manager
fi



