# Run `make` for a list of targets,
# and `make info` to display current configuration.

########## Config

### Common

OPERAND_VERSION ?= 2.x
LC_OPERAND_VERSION = $(shell echo $(OPERAND_VERSION) | tr A-Z a-z)

OPERATOR_VERSION = $(shell sed -n 's/^.*Version.*=.*"\(.*\)".*$$/\1/p' ./version/version.go)
DASH_VERSION = $(shell echo "$(OPERATOR_VERSION)" | sed -n 's/^[0-9\.]*-\([^-+]*\).*$$/-\1/p')

OPERATOR_IMAGE_REPOSITORY ?= quay.io/apicurio
OPERATOR_IMAGE_NAME ?= $(OPERATOR_IMAGE_REPOSITORY)/apicurio-registry-operator
OPERATOR_IMAGE ?= $(OPERATOR_IMAGE_NAME):$(OPERATOR_VERSION)

ADD_LATEST_TAG ?= true

NAMESPACE ?= "apicurio-registry-operator-namespace"
CLIENT ?= kubectl

### Package Manifests

CHANNEL ?= 2.x
IS_DEFAULT_CHANNEL ?= 1
FROM_VERSION = 1.0.0-v2.0.0.final

REPLACES = apicurio-registry-operator.v$(FROM_VERSION)

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
PACKAGE_VERSION = $(OPERATOR_VERSION)-v$(LC_OPERAND_VERSION)

### Bundle

CHANNELS ?= "2.x"
DEFAULT_CHANNEL ?= "2.x"

BUNDLE_IMAGE_NAME ?= $(OPERATOR_IMAGE_NAME)-bundle
BUNDLE_IMAGE ?= $(BUNDLE_IMAGE_NAME):$(PACKAGE_VERSION)

ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS = --channels=$(CHANNELS)
endif

ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL = --default-channel=$(DEFAULT_CHANNEL)
endif

BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

### Catalog

# Version of the catalog image. It is a simple increasing sequence
# of numbers, which must be incremented on each release.
# It is not the same as PACKAGE_VERSION in case there is branching
# in the future
CATALOG_TAG ?= 1
USE_OFFICIAL_PREVIOUS_CATALOG ?= true
# PREVIOUS_CATALOG_TAG ?= latest$(DASH_VERSION)
# TODO ^ after release

CATALOG_IMAGE_NAME = $(OPERATOR_IMAGE_NAME)-catalog
CATALOG_IMAGE ?= $(CATALOG_IMAGE_NAME):$(CATALOG_TAG)$(DASH_VERSION)

ifeq ($(USE_OFFICIAL_PREVIOUS_CATALOG), true)
PREVIOUS_CATALOG_IMAGE = quay.io/apicurio/apicurio-registry-operator-catalog:$(PREVIOUS_CATALOG_TAG)
else
PREVIOUS_CATALOG_IMAGE = $(CATALOG_IMAGE_NAME):$(PREVIOUS_CATALOG_TAG)
endif

ifneq ($(origin PREVIOUS_CATALOG_TAG), undefined)
FROM_INDEX_OPT := --from-index $(PREVIOUS_CATALOG_IMAGE)
endif

### Other



OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)
DATE=$(shell date -Idate)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


########## Info


# Colors!
CC_RED = $(shell echo -e "\033[0;31m")
CC_YELLOW = $(shell echo -e "\033[0;33m")
CC_CYAN = $(shell echo -e "\033[0;36m")
CC_END = $(shell echo -e "\033[0m")

# Parse target documentation from '##' comments
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CC_CYAN)%-30s$(CC_END) %s\n", $$1, $$2}'

.DEFAULT_GOAL = help


.PHONY: info
info:
	@echo "=============================================="
	@echo "Configuration properties:"
	@echo ""
	@echo "$(CC_CYAN)Operator$(CC_END)"
	@echo "OPERATOR_VERSION=$(OPERATOR_VERSION)"
	@echo "DASH_VERSION=$(DASH_VERSION)"
	@echo ""
	@echo "$(CC_CYAN)Operand$(CC_END)"
	@echo "OPERAND_VERSION=$(OPERAND_VERSION)"
	@echo "LC_OPERAND_VERSION=$(LC_OPERAND_VERSION)"
	@echo ""
	@echo "$(CC_CYAN)Bundle$(CC_END)"
	@echo "CHANNELS=$(CHANNELS)"
	@echo "DEFAULT_CHANNEL=$(DEFAULT_CHANNEL)"
	@echo ""
	@echo "$(CC_CYAN)Manifests$(CC_END)"
	@echo "CHANNEL=$(CHANNEL)"
	@echo "IS_DEFAULT_CHANNEL=$(IS_DEFAULT_CHANNEL)"
	@echo "FROM_VERSION=$(FROM_VERSION)"
	@echo "DEFAULT_CHANNEL=$(DEFAULT_CHANNEL)"
	@echo "PACKAGE_MANIFESTS_OPTS=$(PACKAGE_MANIFESTS_OPTS)"
	@echo "PACKAGE_VERSION=$(PACKAGE_VERSION)"
	@echo ""
	@echo "$(CC_CYAN)Tests$(CC_END)"
	@echo "ENVTEST_K8S_VERSION=$(ENVTEST_K8S_VERSION)"
	@echo "=============================================="


########## Tools


# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef


CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: install-controller-gen
install-controller-gen: ## Install controller-gen@v0.4.1
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)


KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: install-kustomize
install-kustomize: ## Install kustomize@v4.5.5
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.5)


YQ_VERSION?="4.9.2"
YQ = $(shell pwd)/bin/yq
.PHONY: install-yq
install-yq: ## Install yq@v4.9.2
	@{ \
		if [ ! -f $(YQ) ]; \
		then \
			mkdir -p $(dir $(YQ)); \
			curl -sSLo $(YQ) "https://github.com/mikefarah/yq/releases/download/v$(YQ_VERSION)/yq_$(OS)_$(ARCH)"; \
			chmod +x $(YQ); \
		fi; \
	}


OPERATOR_SDK_VERSION="1.14.0"
OPERATOR_SDK = $(shell pwd)/bin/operator-sdk
.PHONY: install-operator-sdk
install-operator-sdk: ## Install operator-sdk@v1.14.0
	@{ \
		if [ ! -f $(OPERATOR_SDK) ]; \
		then \
			mkdir -p $(dir $(OPERATOR_SDK)); \
			curl -sSLo $(OPERATOR_SDK) "https://github.com/operator-framework/operator-sdk/releases/download/v$(OPERATOR_SDK_VERSION)/operator-sdk_$(OS)_$(ARCH)"; \
			chmod +x $(OPERATOR_SDK); \
		fi; \
	}


.PHONY: install-antora
install-antora: ## Install antora
ifeq (,$(shell which antora 2> /dev/null))
	@echo "Installing antora using 'sudo npm i -g @antora/cli @antora/site-generator'"
	sudo npm i -g @antora/cli @antora/site-generator
endif


.PHONY: install-opm
OPM = ./bin/opm
install-opm: ## Install opm@v1.19.1
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.19.1/$(OS)-$(ARCH)-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif


########## Targets


.PHONY: all
all: build bundle-build dist ## Perform a full build except the catalog image, which requires bundle to be pushed


.PHONY: all-push
all-push: docker-push bundle-push catalog-build catalog-push ## Build a catalog image and push all images


.PHONY: generate
generate: install-controller-gen ## Generate code
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."


.PHONY: go-fmt
go-fmt: ## Run 'go fmt' against code
	go fmt ./...


.PHONY: go-vet
go-vet: ## Run 'go vet' against code
	go vet ./...


.PHONY: manager
manager: generate go-fmt go-vet ## Build manager binary
	go build -o bin/manager main.go


.PHONY: manifests
manifests: install-controller-gen install-kustomize install-yq ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) rbac:roleName=apicurio-registry-operator-role crd paths="./..." output:crd:artifacts:config=config/crd/resources output:rbac:artifacts:config=config/rbac/resources
	cd config/manager && $(KUSTOMIZE) edit set image REGISTRY_OPERATOR_IMAGE=$(OPERATOR_IMAGE)
	$(YQ) e ".metadata.annotations.createdAt = \"$(DATE)\"" -i "config/manifests/resources/apicurio-registry-operator.clusterserviceversion.yaml"
	$(YQ) e ".metadata.annotations.containerImage = \"$(OPERATOR_IMAGE)\"" -i "config/manifests/resources/apicurio-registry-operator.clusterserviceversion.yaml"
	$(YQ) e ".spec.replaces = \"$(REPLACES)\"" -i "config/manifests/resources/apicurio-registry-operator.clusterserviceversion.yaml"


.PHONY: test
test: generate go-fmt go-vet ## Run unit tests
	go test ./controllers/...


ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
ENVTEST_K8S_VERSION?=1.23
.PHONY: envtest
envtest: generate go-fmt go-vet manifests ## Run integration tests using envtest
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./test/envtest/... -coverprofile cover.out


.PHONY: deploy
deploy: manifests install-kustomize ## Deploy the Operator to a cluster using $CLIENT (kubectl)
	cd config/manager && $(KUSTOMIZE) edit set image REGISTRY_OPERATOR_IMAGE=${OPERATOR_IMAGE}
	cd config/default && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
	$(CLIENT) create namespace $(NAMESPACE) || true
	$(KUSTOMIZE) build config/default | $(CLIENT) apply -f - -n $(NAMESPACE)


.PHONY: undeploy
undeploy: install-kustomize ## Un-deploy the Operator from a cluster using $CLIENT (kubectl)
	$(KUSTOMIZE) build config/default | $(CLIENT) delete --ignore-not-found=true -f -


.PHONY: docker-build
docker-build: test ## Build Operator image
	docker build -t ${OPERATOR_IMAGE} .


.PHONY: docker-push
docker-push: ## Push Operator image
ifeq ($(ADD_LATEST_TAG),true)
	docker tag $(OPERATOR_IMAGE) $(OPERATOR_IMAGE_NAME):latest$(DASH_VERSION)
	docker push $(OPERATOR_IMAGE_NAME):latest$(DASH_VERSION)
endif
	docker push ${OPERATOR_IMAGE}


.PHONY: build
build: test envtest manager docker-build ## Build Operator image


.PHONY: bundle
bundle: manifests install-operator-sdk install-yq ## Generate bundle manifests and metadata
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(PACKAGE_VERSION) $(BUNDLE_METADATA_OPTS)
	# Workaround for https://github.com/operator-framework/operator-lifecycle-manager/issues/1608
	# See https://github.com/operator-framework/operator-lifecycle-manager/issues/952#issuecomment-639657949
	$(YQ) e ".spec.install.spec.deployments[0].name = .spec.install.spec.deployments[0].name + \"-v$(OPERATOR_VERSION)\"" -i "bundle/manifests/apicurio-registry-operator.clusterserviceversion.yaml"
	$(OPERATOR_SDK) bundle validate ./bundle


.PHONY: bundle-build
bundle-build: bundle ## Build bundle image
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMAGE) .


.PHONY: bundle-push
bundle-push: ## Push bundle image
ifeq ($(ADD_LATEST_TAG),true)
	docker tag $(BUNDLE_IMAGE) $(BUNDLE_IMAGE_NAME):latest$(DASH_VERSION)
	docker push $(BUNDLE_IMAGE_NAME):latest$(DASH_VERSION)
endif
	docker push $(BUNDLE_IMAGE)


.PHONY: packagemanifests
packagemanifests: manifests install-operator-sdk install-yq ## Generate package manifests
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate packagemanifests -q --version $(PACKAGE_VERSION) $(PACKAGE_MANIFESTS_OPTS)
	# Workaround for https://github.com/operator-framework/operator-lifecycle-manager/issues/1608
	# See https://github.com/operator-framework/operator-lifecycle-manager/issues/952#issuecomment-639657949
	$(YQ) e ".spec.install.spec.deployments[0].name = .spec.install.spec.deployments[0].name + \"-v$(OPERATOR_VERSION)\"" -i "packagemanifests/$(PACKAGE_VERSION)/apicurio-registry-operator.clusterserviceversion.yaml"


.PHONY: docs
docs: install-antora ## Build documentation
	cd ./docs && antora local-test-playbook.yml
	@echo "The docs are available at:"
	@echo "file:$(shell pwd)/docs/target/dist/index.html"


.PHONY: dist
dist: install-kustomize docs ## Generate distribution archive
	mkdir -p dist
	cp -rt ./dist ./dist-base/*
	cp -t ./dist ./LICENSE
	# Examples
	cp -t ./dist/examples ./config/examples/resources/*
	cp -rt ./dist/examples ./docs/modules/ROOT/examples/*
	# Docs
	cp -rt ./dist ./docs/target/dist && mv ./dist/dist ./dist/docs
	# Install
	cd config/manager && $(KUSTOMIZE) edit set image REGISTRY_OPERATOR_IMAGE=$(OPERATOR_IMAGE)
	$(KUSTOMIZE) build config/default/ > ./dist/install.yaml
	$(KUSTOMIZE) build config/default/ > ./docs/resources/install.yaml
	# Archive
	tar -zcf apicurio-registry-operator-$(PACKAGE_VERSION).tar.gz -C ./dist .


.PHONY: clean
clean: ## Remove temporary and generated files
	@rm apicurio-registry-operator-$(PACKAGE_VERSION).tar.gz cover.out 2>/dev/null || true
	@rm -r bin build bundle bundle_tmp* dist docs/target testbin 2>/dev/null || true


.PHONY: catalog-build
catalog-build: install-opm ## Build the catalog image
	@echo "Note: You need to build and push your bundle image before building a catalog image."
	@echo "Run 'make bundle-build bundle-push' t do this."
	# TODO: Remove the first bundle in the list after we can start using previous catalog images
	$(OPM) index add --container-tool docker --bundles quay.io/apicurio/apicurio-registry-operator-bundle:1.0.0-v2.0.0.final,$(BUNDLE_IMAGE) --tag $(CATALOG_IMAGE) $(FROM_INDEX_OPT)


.PHONY: catalog-push
catalog-push: ## Push the catalog image
ifeq ($(ADD_LATEST_TAG),true)
	docker tag $(CATALOG_IMAGE) $(CATALOG_IMAGE_NAME):latest$(DASH_VERSION)
	docker push $(CATALOG_IMAGE_NAME):latest$(DASH_VERSION)
endif
	docker push $(CATALOG_IMAGE)
