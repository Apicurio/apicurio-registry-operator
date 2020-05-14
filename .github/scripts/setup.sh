#!/bin/bash
set -e

echo "Set up golang"
export GO111MODULE="on"

echo "Install Operator SDK"
export OPERATOR_SDK_VERSION="0.17.0"
curl -fsSL "https://github.com/operator-framework/operator-sdk/releases/download/v$OPERATOR_SDK_VERSION/operator-sdk-v$OPERATOR_SDK_VERSION-x86_64-linux-gnu" >operator-sdk
chmod +x operator-sdk
sudo mv operator-sdk /usr/local/bin/operator-sdk

echo "Clean cache"
sudo apt-get clean

