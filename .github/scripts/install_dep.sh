#!/bin/bash
set -e

echo "Set up golang"
export GO111MODULE='on'

echo "Install operator sdk"
curl -fsSL https://github.com/operator-framework/operator-sdk/releases/download/v0.17.0/operator-sdk-v0.17.0-x86_64-linux-gnu > operator-sdk
chmod +x operator-sdk
sudo mv operator-sdk /usr/local/bin/operator-sdk

echo "Clean cache"
sudo apt-get clean
