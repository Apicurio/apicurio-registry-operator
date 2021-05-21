#!/bin/bash
set -e

echo "Set up golang"
export GO111MODULE="on"

echo "Install Operator SDK"
export OPERATOR_SDK_VERSION="1.4.2"
curl -fsSL "https://github.com/operator-framework/operator-sdk/releases/download/v$OPERATOR_SDK_VERSION/operator-sdk_linux_amd64" >operator-sdk
chmod +x operator-sdk
sudo mv operator-sdk /usr/local/bin/operator-sdk

echo "Clean cache"
sudo apt-get clean

echo "Install Antora"
sudo npm i -g @antora/cli @antora/site-generator-default
antora -v

echo "Install 'yq' v4.x"
export YQ_VERSION="v4.9.2"
curl -fsSL "https://github.com/mikefarah/yq/releases/download/$YQ_VERSION/yq_linux_amd64" >yq
chmod +x yq
sudo mv yq /usr/local/bin/yq
