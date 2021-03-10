# define HELP
# endef
#
# export HELP
# .PHONY: help
# help:
# 	@echo "$$HELP"

# Parse target documentation from '##' comments
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

### Config

OPERATOR_VERSION ?= 1.0.0-dev

OPERAND_VERSION ?= 2.0.0-SNAPSHOT

LC_OPERAND_VERSION = $(shell echo $(OPERAND_VERSION) | tr A-Z a-z)

CHANNELS = "apicurio-registry-2.x" # TODO alpha?

ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

DEFAULT_CHANNEL = "apicurio-registry-2.x"

ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif

BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

OPERATOR_IMAGE_REPOSITORY ?= "quay.io/apicurio"
OPERATOR_IMAGE_NAME ?= "$(OPERATOR_IMAGE_REPOSITORY)/apicurio-registry-operator"
OPERATOR_IMAGE ?= "$(OPERATOR_IMAGE_NAME):$(OPERATOR_VERSION)"

BUNDLE_IMAGE_NAME ?= "$(OPERATOR_IMAGE_NAME)-bundle"
BUNDLE_IMAGE ?= "$(BUNDLE_IMAGE_NAME):$(OPERATOR_VERSION)"

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

### Env

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

### Tools

OPERATOR_SDK = $(shell which operator-sdk)

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen@v0.4.1 using 'go get'
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)


KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize@v3.8.7 using 'go get'
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)


# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef


###

all: manager ## Run manager target


ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests ## Run tests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out


manager: generate fmt vet ## Build manager binary
	go build -o bin/manager main.go


run: generate fmt vet manifests ## ??? Run against the configured Kubernetes cluster in ~/.kube/config
	go run ./main.go


install: manifests kustomize ## Install CRDs into a cluster
	$(KUSTOMIZE) build config/crd | kubectl apply -f -


uninstall: manifests kustomize ## Uninstall CRDs from a cluster
	$(KUSTOMIZE) build config/crd | kubectl delete -f -


deploy: manifests kustomize ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && $(KUSTOMIZE) edit set image controller=${OPERATOR_IMAGE}
	$(KUSTOMIZE) build config/default | kubectl apply -f -


undeploy: ## Undeploy controller from the configured Kubernetes cluster in ~/.kube/config
	$(KUSTOMIZE) build config/default | kubectl delete -f -


manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases


fmt: ## Run 'go fmt' against code
	go fmt ./...


vet: ## Run 'go vet' against code
	go vet ./...


generate: controller-gen ## ??? Generate code
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."


docker-build: test ## Build operator docker image
	docker build -t ${OPERATOR_IMAGE} .


docker-push: ## Push operator docker image
	docker push ${OPERATOR_IMAGE}





.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(OPERATOR_IMAGE)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle


.PHONY: bundle-build
bundle-build: ## Build the bundle image
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMAGE) .


# Options for "packagemanifests".

CHANNEL ?= "apicurio-registry-2.x"
IS_DEFAULT_CHANNEL ?= 1

ifneq ($(origin FROM_VERSION), undefined)
PACKAGE_FROM_VERSION := --from-version=$(FROM_VERSION)
endif

ifneq ($(origin CHANNEL), undefined)
PACKAGE_CHANNELS := --channel=$(CHANNEL)
endif

ifeq ($(IS_DEFAULT_CHANNEL), 1)
PACKAGE_IS_DEFAULT_CHANNEL := --default-channel
endif

PACKAGE_MANIFESTS_OPTS ?= $(PACKAGE_FROM_VERSION) $(PACKAGE_CHANNELS) $(PACKAGE_IS_DEFAULT_CHANNEL)

PACKAGE_VERSION = $(OPERATOR_VERSION)-$(LC_OPERAND_VERSION)

packagemanifests: kustomize manifests ## Generate package manifests.
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(OPERATOR_IMAGE)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate packagemanifests -q --version $(PACKAGE_VERSION) $(PACKAGE_MANIFESTS_OPTS)


.PHONY: foo
foo:
	echo $(LC_OPERAND_VERSION)
