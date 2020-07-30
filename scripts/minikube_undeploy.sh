#!/bin/sh




  kubectl config set-context --current --namespace=$OPERATOR_NAMESPACE

  kubectl delete -f ./deploy/crds/

  kubectl delete deployment apicurio-registry-operator
  kubectl delete CustomResourceDefinition apicurioregistries.apicur.io
  kubectl delete RoleBinding apicurio-registry-operator
  kubectl delete Role apicurio-registry-operator
  kubectl delete ClusterRoleBinding apicurio-registry-operator
  kubectl delete ClusterRole apicurio-registry-operator
  kubectl delete ServiceAccount apicurio-registry-operator
