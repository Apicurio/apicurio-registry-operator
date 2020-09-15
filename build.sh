#!/bin/bash

help() {
  echo "Help:"
  echo "Apicurio Registry Operator build tool"
  echo "Note: Run this script from the root dir of the project."
  echo -e "\n$0 [command] [parameters]..."
  echo -e "\nCommands: "
  echo "  build"
  echo "  help"
  echo "  mkdeploy"
  echo "  mkundeploy"
  echo "  push"
  echo -e "\nParameters:"
  echo "  -r|--repository [repository] Operator image repository"
  echo "  -n|--namespace [namespace] Namespace where the operator is deployed"
  echo "  --cr [file] Path to a file with 'ApicurioRegistry' custom resource to be deployed"
  echo "  --nocr Do not deploy default 'ApicurioRegistry' custom resource"
  echo "  --crname [name] Name of the 'ApicurioRegistry' custom resource (e.g. for mkundeploy), default is 'example-apicurioregistry'"
  echo "  --latest Also push the image with the 'latest' tag"
  exit 1
}

error() {
  echo -e "Error: $1\n"
  help
}

require() {
  if [[ -z "$1" ]]; then
    error "$2"
  fi
}

init_image() {
  require "$OPERATOR_IMAGE_REPOSITORY" "Parameter '-r' is required."
  VERSION=$(sed -n 's/^.*Version.*=.*"\(.*\)".*$/\1/p' ./version/version.go)
  DASH_VERSION_RELEASE=$(echo "$VERSION" | sed -n 's/^[0-9\.]*-\([^-+]*\).*$/-\1/p')
  require "$VERSION" "Could not read project version."

  # Operator
  OPERATOR_IMAGE_NAME="$OPERATOR_IMAGE_REPOSITORY/apicurio-registry-operator"
  OPERATOR_IMAGE="$OPERATOR_IMAGE_NAME:$VERSION"
  OPERATOR_IMAGE_LATEST="$OPERATOR_IMAGE_NAME:latest$DASH_VERSION_RELEASE"

  # Metadata
  METADATA_IMAGE_NAME="$OPERATOR_IMAGE_NAME-metadata"
  METADATA_IMAGE="$METADATA_IMAGE_NAME:$VERSION"
  METADATA_IMAGE_LATEST="$METADATA_IMAGE_NAME:latest$DASH_VERSION_RELEASE"

  # Registry
  REGISTRY_IMAGE_MEM="docker.io/apicurio/apicurio-registry-mem:1.2.3.Final"
  REGISTRY_IMAGE_KAFKA="docker.io/apicurio/apicurio-registry-kafka:1.2.3.Final"
  REGISTRY_IMAGE_STREAMS="docker.io/apicurio/apicurio-registry-streams:1.2.3.Final"
  REGISTRY_IMAGE_JPA="docker.io/apicurio/apicurio-registry-jpa:1.2.3.Final"
  REGISTRY_IMAGE_INFINISPAN="docker.io/apicurio/apicurio-registry-infinispan:1.2.3.Final"

}

replace() {
  init_image
  sed -i "s|{OPERATOR_IMAGE}|$OPERATOR_IMAGE # replaced {OPERATOR_IMAGE}|g" ./deploy/operator.yaml
  sed -i "s|{REGISTRY_IMAGE_MEM}|$REGISTRY_IMAGE_MEM # replaced {REGISTRY_IMAGE_MEM}|g" ./deploy/operator.yaml
  sed -i "s|{REGISTRY_IMAGE_KAFKA}|$REGISTRY_IMAGE_KAFKA # replaced {REGISTRY_IMAGE_KAFKA}|g" ./deploy/operator.yaml
  sed -i "s|{REGISTRY_IMAGE_STREAMS}|$REGISTRY_IMAGE_STREAMS # replaced {REGISTRY_IMAGE_STREAMS}|g" ./deploy/operator.yaml
  sed -i "s|{REGISTRY_IMAGE_JPA}|$REGISTRY_IMAGE_JPA # replaced {REGISTRY_IMAGE_JPA}|g" ./deploy/operator.yaml
  sed -i "s|{REGISTRY_IMAGE_INFINISPAN}|$REGISTRY_IMAGE_INFINISPAN # replaced {REGISTRY_IMAGE_INFINISPAN}|g" ./deploy/operator.yaml

}

unreplace() {
  sed -i "s|$OPERATOR_IMAGE # replaced {OPERATOR_IMAGE}|{OPERATOR_IMAGE}|g" ./deploy/operator.yaml
  sed -i "s|$REGISTRY_IMAGE_MEM # replaced {REGISTRY_IMAGE_MEM}|{REGISTRY_IMAGE_MEM}|g" ./deploy/operator.yaml
  sed -i "s|$REGISTRY_IMAGE_KAFKA # replaced {REGISTRY_IMAGE_KAFKA}|{REGISTRY_IMAGE_KAFKA}|g" ./deploy/operator.yaml
  sed -i "s|$REGISTRY_IMAGE_STREAMS # replaced {REGISTRY_IMAGE_STREAMS}|{REGISTRY_IMAGE_STREAMS}|g" ./deploy/operator.yaml
  sed -i "s|$REGISTRY_IMAGE_JPA # replaced {REGISTRY_IMAGE_JPA}|{REGISTRY_IMAGE_JPA}|g" ./deploy/operator.yaml
  sed -i "s|$REGISTRY_IMAGE_INFINISPAN # replaced {REGISTRY_IMAGE_INFINISPAN}|{REGISTRY_IMAGE_INFINISPAN}|g" ./deploy/operator.yaml
}

gen_csv() {
  # Generate dev CRDs, alpha channel
  if [ -d "./deploy/olm-catalog/apicurio-registry/$VERSION" ]; then
    operator-sdk generate csv \
      --csv-channel alpha \
      --update-crds \
      --csv-version "$VERSION" \
      --operator-name apicurio-registry \
      --verbose \
      --from-version 0.0.0-template \
      --make-manifests=false

    CSV_PATH="./deploy/olm-catalog/apicurio-registry/$VERSION/apicurio-registry.v$VERSION.clusterserviceversion.yaml"
    PACKAGE_PATH="./deploy/olm-catalog/apicurio-registry/apicurio-registry-operator.package.yaml"

    PREVIOUS_VERSION_ALPHA=$(sed -n 's/^ *currentCSV:.*# alpha; replaces \([^ ]*\)$/\1/p' "$PACKAGE_PATH")
    require "$PREVIOUS_VERSION_ALPHA" "Could not determine previous CSV version."
    sed -i "s/replaces: *apicurio-registry\.v0\.0\.0-template/replaces: apicurio-registry.v$PREVIOUS_VERSION_ALPHA/" "$CSV_PATH"

    CREATED_AT=$(date -Idate)
    sed -i "s/createdAt: .*/createdAt: \"$CREATED_AT\"/" "$CSV_PATH"

    sed -i "s|containerImage: .*|containerImage: \"$OPERATOR_IMAGE\"|" "$CSV_PATH"

    # sed -i "s/\(^ *currentCSV: *apicurio-registry\.v\)\([^ ]*\)\( *# *alpha.*$\)/\1$VERSION\3/" "$PACKAGE_PATH"

    echo "Warning: Make sure generated CSV do not contain your private dev changes before commiting."
    echo "Warning: If you want to create CSV for release, rename the generated 'dev' one, replace tags with SHA using 'skopeo' and add 'related images' section."
  fi
}

build() {
  replace
  operator-sdk generate k8s
  operator-sdk generate crds

  gen_csv

  # Operator
  operator-sdk build "$OPERATOR_IMAGE"
  docker tag "$OPERATOR_IMAGE" "$OPERATOR_IMAGE_LATEST" # Tag as latest

  # Metadata
  docker build -t "$METADATA_IMAGE" "./deploy/olm-catalog/"
  docker tag "$METADATA_IMAGE" "$METADATA_IMAGE_LATEST" # Tag as latest

  compile_qs_yaml
  #unreplace
}

minikube_deploy_cr() {
  require "$OPERATOR_NAMESPACE" "Argument -n or --namespace is required."
  if [[ -z "$CR_PATH" ]]; then
    if [[ -z "$NO_DEFAULT_CR" ]]; then
      kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_cr.yaml -n "$OPERATOR_NAMESPACE"
    fi
  else
    kubectl create -f "$CR_PATH" -n "$OPERATOR_NAMESPACE"
  fi
}

minikube_deploy() {
  require "$OPERATOR_NAMESPACE" "Argument -n or --namespace is required."
  replace
  kubectl create -f ./deploy/service_account.yaml -n "$OPERATOR_NAMESPACE"
  kubectl create -f ./deploy/role.yaml  -n "$OPERATOR_NAMESPACE"
  kubectl create -f ./deploy/role_binding.yaml  -n "$OPERATOR_NAMESPACE"
  kubectl create -f ./deploy/cluster_role.yaml
  cat ./deploy/cluster_role_binding.yaml | sed "s/{NAMESPACE}/$OPERATOR_NAMESPACE # replaced {NAMESPACE}/g" | kubectl apply -f -
  kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_crd.yaml
  kubectl create -f ./deploy/operator.yaml -n "$OPERATOR_NAMESPACE"
  minikube_deploy_cr
  kubectl get deployments -n "$OPERATOR_NAMESPACE"
  unreplace
}

compile_qs_yaml() {
  FILE="./docs/resources/install.yaml"
  echo "Warning: Make sure generated files like '$FILE' do not contain your private dev changes (e.g. image references) before commiting."
  if [ -f "$FILE" ]; then
    rm "$FILE"
  fi
  echo -e "\n---"  >> "$FILE" && cat ./deploy/service_account.yaml >> "$FILE"
  echo -e "\n---"  >> "$FILE" && cat ./deploy/role.yaml >> "$FILE"
  echo -e "\n---"  >> "$FILE" && cat ./deploy/role_binding.yaml >> "$FILE"
  echo -e "\n---"  >> "$FILE" && cat ./deploy/cluster_role.yaml >> "$FILE"
  echo -e "\n---"  >> "$FILE" && cat ./deploy/cluster_role_binding.yaml >> "$FILE"
  echo -e "\n---"  >> "$FILE" && cat ./deploy/crds/apicur.io_apicurioregistries_crd.yaml >> "$FILE"
  echo -e "\n---"  >> "$FILE" && cat ./deploy/operator.yaml >> "$FILE"
  echo ""  >> "$FILE"
}

minikube_undeploy() {
  require "$OPERATOR_NAMESPACE" "Argument -n or --namespace is required."
  #kubectl delete ApicurioRegistry "$CR_NAME"
  kubectl delete deployment apicurio-registry-operator -n "$OPERATOR_NAMESPACE"
  kubectl delete CustomResourceDefinition apicurioregistries.apicur.io
  kubectl delete RoleBinding apicurio-registry-operator -n "$OPERATOR_NAMESPACE"
  kubectl delete Role apicurio-registry-operator -n "$OPERATOR_NAMESPACE"
  kubectl delete ClusterRoleBinding apicurio-registry-operator
  kubectl delete ClusterRole apicurio-registry-operator
  kubectl delete ServiceAccount apicurio-registry-operator -n "$OPERATOR_NAMESPACE"
}

push() {
  init_image
  docker push "$OPERATOR_IMAGE"
  docker push "$METADATA_IMAGE" # Metadata

  if [[ -n "$PUSH_LATEST" ]]; then
    docker push "$OPERATOR_IMAGE_LATEST"
    docker push "$METADATA_IMAGE_LATEST" # Metadata
  fi
}

if [ ! -f "./version/version.go" ]; then
    echo "Please run this script from the repository root."
    exit 1
fi

TARGET="$1"
shift

while [[ "$#" -gt 0 ]]; do
  case $1 in
  -r | --repository)
    OPERATOR_IMAGE_REPOSITORY="$2"
    shift
    ;;
  -n | --namespace)
    OPERATOR_NAMESPACE="$2"
    shift
    ;;
  --cr)
    CR_PATH="$2"
    shift
    ;;
  --nocr)
    NO_DEFAULT_CR="true"
    shift
    ;;
  --crname)
    CR_NAME="$2"
    shift
    ;;
  --latest)
    PUSH_LATEST="true"
    shift
    ;;
  *)
    echo -e "Unknown parameter: '$1'.\n"
    help
    ;;
  esac
  shift
done

if [[ -z "$CR_NAME" ]]; then
  CR_NAME="example-apicurioregistry"
fi

case "$TARGET" in
build) build ;;
mkdeploy) minikube_deploy ;;
mkundeploy) minikube_undeploy ;;
push) push ;;
help) help ;;
*)
  echo -e "Unknown command: '$TARGET'.\n"
  help
  ;;
esac
