#!/bin/bash

# Adapted for Apicurio from https://github.com/keycloak/keycloak-operator/blob/master/hack/modify_etc_hosts.sh

# The main part of this script has been downloaded from: https://gist.github.com/jacobtomlinson/4b835d807ebcea73c6c8f602613803d4

HOST=$1
MINIKUBE_IP=$2

help() {
  echo "Help:"
  echo "Adds an entry to /etc/hosts to configure ingress access to Minikube"
  echo "Usage: ./modify_etc_hosts.sh <host> [<minikube ip>]"
  exit 1
}

if [ -z "$HOST" ]; then
  echo "Error: Ingress address (host) is not set"
  help
fi

if [ -z "$MINIKUBE_IP" ]; then
  echo "Make sure Minikube is running"
  MINIKUBE_IP=$(minikube ip || help)
fi

HOSTS_ENTRY="$MINIKUBE_IP $HOST"

if grep -Fq "$HOST" /etc/hosts >/dev/null; then
  echo "s/^.*$HOST.*$/$HOSTS_ENTRY/"
  sudo sed -i "s/^.*$HOST.*$/$HOSTS_ENTRY/" /etc/hosts
  echo "Updated hosts entry:"
  sudo cat /etc/hosts
else
  echo "$HOSTS_ENTRY" | sudo tee -a /etc/hosts
  echo "Added hosts entry:"
  sudo cat /etc/hosts
fi
