#!/bin/bash
set -e

echo "Set up golang"
export GO111MODULE="on"

echo "Clean cache"
sudo apt-get clean
