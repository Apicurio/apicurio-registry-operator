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

require_command() {
  if ! command -v "$1" &>/dev/null; then
    error "Required command '$1' is not available. $2"
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

}

replace() {
  init_image
  sed -i "s|{OPERATOR_IMAGE}|$OPERATOR_IMAGE # replaced {OPERATOR_IMAGE}|g" ./deploy/operator.yaml
}

unreplace() {
  sed -i "s|$OPERATOR_IMAGE # replaced {OPERATOR_IMAGE}|{OPERATOR_IMAGE}|g" ./deploy/operator.yaml
}

gen_csv() {
  # Generate dev CRDs, alpha channel
  if [ -d "./deploy/olm-catalog/apicurio-registry/$VERSION" ]; then

    require_command yq "You can find the installation instructions at https://mikefarah.gitbook.io/yq/ or see './.github/scripts/setup.sh'"

    operator-sdk generate csv \
      --csv-channel alpha \
      --update-crds \
      --csv-version "$VERSION" \
      --operator-name apicurio-registry \
      --verbose \
      --from-version 0.0.0-template \
      --make-manifests=false

    if [ ! $? ]; then
      error "Could not re-generate CSV."
    fi

    CSV_TEMPLATE_PATH="./deploy/olm-catalog/apicurio-registry/0.0.0-template/apicurio-registry.v0.0.0-template.clusterserviceversion.yaml"
    CSV_PATH="./deploy/olm-catalog/apicurio-registry/$VERSION/apicurio-registry.v$VERSION.clusterserviceversion.yaml"
    PACKAGE_PATH="./deploy/olm-catalog/apicurio-registry/apicurio-registry-operator.package.yaml"
    PREVIOUS_VERSION_ALPHA=$(sed -n 's/^ *currentCSV:.*# alpha; replaces \([^ ]*\)$/\1/p' "$PACKAGE_PATH")
    require "$PREVIOUS_VERSION_ALPHA" "Could not determine previous CSV version."

    # Copy specDescriptors from template, this has to be done explicitly
    yq r "$CSV_TEMPLATE_PATH" "spec.customresourcedefinitions.owned[0].specDescriptors" |
      yq p - "spec.customresourcedefinitions.owned[0].specDescriptors" |
      yq m -i -P "$CSV_PATH" -

    # Update the relatedImages section
    yq r "$CSV_TEMPLATE_PATH" "spec.relatedImages" |
      yq p - "spec.relatedImages" |
      yq m -i -P "$CSV_PATH" -

    _IMAGE=$(yq r "$CSV_PATH" "spec.install.spec.deployments[0].spec.template.spec.containers[0].image")
    yq w -i "$CSV_PATH" "spec.relatedImages[name==apicurio-registry-operator].image" "$_IMAGE"
    _IMAGE=$(yq r "$CSV_PATH" "spec.install.spec.deployments[0].spec.template.spec.containers[0].env[name==REGISTRY_IMAGE_MEM].value")
    yq w -i "$CSV_PATH" "spec.relatedImages[name==apicurio-registry-mem].image" "$_IMAGE"
    _IMAGE=$(yq r "$CSV_PATH" "spec.install.spec.deployments[0].spec.template.spec.containers[0].env[name==REGISTRY_IMAGE_KAFKA].value")
    yq w -i "$CSV_PATH" "spec.relatedImages[name==apicurio-registry-kafka].image" "$_IMAGE"
    _IMAGE=$(yq r "$CSV_PATH" "spec.install.spec.deployments[0].spec.template.spec.containers[0].env[name==REGISTRY_IMAGE_STREAMS].value")
    yq w -i "$CSV_PATH" "spec.relatedImages[name==apicurio-registry-streams].image" "$_IMAGE"
    _IMAGE=$(yq r "$CSV_PATH" "spec.install.spec.deployments[0].spec.template.spec.containers[0].env[name==REGISTRY_IMAGE_JPA].value")
    yq w -i "$CSV_PATH" "spec.relatedImages[name==apicurio-registry-jpa].image" "$_IMAGE"
    _IMAGE=$(yq r "$CSV_PATH" "spec.install.spec.deployments[0].spec.template.spec.containers[0].env[name==REGISTRY_IMAGE_INFINISPAN].value")
    yq w -i "$CSV_PATH" "spec.relatedImages[name==apicurio-registry-infinispan].image" "$_IMAGE"

    # Update the 'replaces' field
    yq w -i -P "$CSV_PATH" "spec.replaces" "apicurio-registry.v$PREVIOUS_VERSION_ALPHA"

    # Update the 'createdAt' field
    CREATED_AT=$(date -Idate)
    yq w -i -P "$CSV_PATH" "metadata.annotations.createdAt" "$CREATED_AT"

    # Update the 'containerImage' field
    yq w -i -P "$CSV_PATH" "metadata.annotations.containerImage" "$OPERATOR_IMAGE"

    echo "⚠️ Warning: Make sure generated CSV does not contain your private dev changes before commiting."
    echo "⚠️ Warning: If you want to create CSV for release, rename the generated 'dev' one and replace tags with SHA using 'skopeo' (also in the 'relatedImages' section)."
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
  kubectl create -f ./deploy/role.yaml -n "$OPERATOR_NAMESPACE"
  kubectl create -f ./deploy/role_binding.yaml -n "$OPERATOR_NAMESPACE"
  kubectl create -f ./deploy/cluster_role.yaml
  cat ./deploy/cluster_role_binding.yaml | sed "s/{NAMESPACE}/$OPERATOR_NAMESPACE # replaced {NAMESPACE}/g" | kubectl apply -f -
  kubectl create -f ./deploy/crds/apicur.io_apicurioregistries_crd.yaml
  kubectl create -f ./deploy/operator.yaml -n "$OPERATOR_NAMESPACE"
  minikube_deploy_cr
  kubectl get deployments -n "$OPERATOR_NAMESPACE"
  unreplace
}

compile_qs_yaml() {
  FILE="./docs/resources/install-dev.yaml"
  echo "⚠️ Warning: Make sure generated files like '$FILE' do not contain your private dev changes (e.g. image references) before commiting."
  if [ -f "$FILE" ]; then
    rm "$FILE"
  fi
  echo -e "\n---" >>"$FILE" && cat ./deploy/service_account.yaml >>"$FILE"
  echo -e "\n---" >>"$FILE" && cat ./deploy/role.yaml >>"$FILE"
  echo -e "\n---" >>"$FILE" && cat ./deploy/role_binding.yaml >>"$FILE"
  echo -e "\n---" >>"$FILE" && cat ./deploy/cluster_role.yaml >>"$FILE"
  echo -e "\n---" >>"$FILE" && cat ./deploy/cluster_role_binding.yaml >>"$FILE"
  echo -e "\n---" >>"$FILE" && cat ./deploy/crds/apicur.io_apicurioregistries_crd.yaml >>"$FILE"
  echo -e "\n---" >>"$FILE" && cat ./deploy/operator.yaml >>"$FILE"
  echo "" >>"$FILE"
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
