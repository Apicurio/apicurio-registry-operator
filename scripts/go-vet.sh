#!/bin/sh

source ./scripts/go-mod-env.sh

if [[ -z ${CI} ]]; then
    ./scripts/go-mod.sh
    ./scripts/go-sdk-gen.sh
fi
go vet ./...
