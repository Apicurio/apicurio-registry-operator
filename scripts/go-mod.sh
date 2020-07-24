#!/bin/sh

source ./scripts/go-mod-env.sh

echo Reset vendor directory

if [[ -z ${CI} ]]; then
    go mod tidy
    go mod vendor
else
    go mod vendor -v
fi
