#!/bin/sh

   echo $OPERATOR_NAMESPACE
   echo $OPERATOR_IMAGE_REPOSITORY


  VERSION=$(sed -n 's/^.*Version.*=.*"\(.*\)".*$/\1/p' ./version/version.go)
  DASH_VERSION_RELEASE=$(echo "$VERSION" | sed -n 's/^[0-9\.]*-\([^-+]*\).*$/-\1/p')
  OPERATOR_IMAGE_NAME="$OPERATOR_IMAGE_REPOSITORY/apicurio-registry-operator"
  OPERATOR_IMAGE="$OPERATOR_IMAGE_NAME:$VERSION"
  echo $OPERATOR_IMAGE
  sed -i "s|{OPERATOR_IMAGE}|$OPERATOR_IMAGE # replaced {OPERATOR_IMAGE}|g" ./deploy/operator.yaml



  kubectl create -f ./deploy/service_account.yaml
  kubectl create -f ./deploy/role.yaml
  kubectl create -f ./deploy/role_binding.yaml
  kubectl create -f ./deploy/cluster_role.yaml
  cat ./deploy/cluster_role_binding.yaml | sed "s/{NAMESPACE}/$OPERATOR_NAMESPACE # replaced {NAMESPACE}/g" | kubectl apply -f -
  kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_crd.yaml
  kubectl create -f ./deploy/operator.yaml


   if [[ -z "$CR_PATH" ]]; then
    if [[ -z "$NO_DEFAULT_CR" ]]; then
      kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_cr.yaml
    fi
  else
    kubectl create -f "$CR_PATH"
  fi

 kubectl get deployments

 sed -i "s|$OPERATOR_IMAGE # replaced {OPERATOR_IMAGE}|{OPERATOR_IMAGE}|g" ./deploy/operator.yaml
