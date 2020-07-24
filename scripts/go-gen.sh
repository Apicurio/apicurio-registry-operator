#!/bin/sh

source ./scripts/go-mod-env.sh

go generate -mod=vendor ./...
