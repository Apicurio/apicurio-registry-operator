#!/bin/sh

export GOFLAGS=-mod=vendor

pwdPath=$(pwd -P 2>/dev/null || env PWD= pwd)
goPath=$(go env GOPATH)
cd "${goPath}" || exit
goPath=$(pwd -P 2>/dev/null || env PWD= pwd)
cd "${pwdPath}" || exit
if [ "${pwdPath#"$goPath"}" != "${pwdPath}" ]; then
  export GO111MODULE=on
fi
