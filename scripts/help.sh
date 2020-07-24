#!/usr/bin/env bash


  echo "Help:"
  echo "Apicurio Registry Operator build tool"
  echo "Note: Run this make targets from the root dir of the project."
  echo -e "\n$0 [command] [parameters]..."
  echo -e "\nCommands: "
  echo "  make build"
  echo "  make deploy"
  echo "  make undeploy"
  echo "  make push"
  echo -e "\nParameters:"
  echo "  OPERATOR_IMAGE_REPOSITORY=quay.io/apicurio  Operator image repository"
  echo "  OPERATOR_NAMESPACE=[Namespace] where the operator is deployed"
  echo "  CR_PATH= [file] Path to a file with 'ApicurioRegistry' custom resource to be deployed ex : docs/resources/example-cr/in-memory.yaml"
  echo "  docker push path of complete operator image ex : docker push quay.io/apicurio/apicurio-registry-operator:0.0.4-dev"

