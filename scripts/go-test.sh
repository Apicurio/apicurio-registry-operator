#!/bin/sh
source ./scripts/go-mod-env.sh

if [[ -z ${CI} ]]; then
    ./scripts/go-vet.sh
    ./scripts/go-fmt.sh

fi
go test ./...
