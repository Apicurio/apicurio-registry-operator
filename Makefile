# Parse target documentation from '##' comments
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL = help

########## Config

OPERAND_VERSION ?= 2.x
LC_OPERAND_VERSION = $(shell echo $(OPERAND_VERSION) | tr A-Z a-z)

OPERATOR_VERSION = $(shell sed -n 's/^.*Version.*=.*"\(.*\)".*$$/\1/p' ./version/version.go)
DASH_VERSION = $(shell echo "$(OPERATOR_VERSION)" | sed -n 's/^[0-9\.]*-\([^-+]*\).*$$/-\1/p')

### Bundle

CHANNELS = "2.x"
DEFAULT_CHANNEL = "2.x"

ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS = --channels=$(CHANNELS)
endif

ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL = --default-channel=$(DEFAULT_CHANNEL)
endif

BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

### Package Manifests

CHANNEL ?= "2.x"
IS_DEFAULT_CHANNEL ?= 1
FROM_VERSION = 1.0.0-v2.1.5.final

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

###

OPERATOR_IMAGE_REPOSITORY ?= quay.io/apicurio

OPERATOR_IMAGE_NAME = $(OPERATOR_IMAGE_REPOSITORY)/apicurio-registry-operator
OPERATOR_IMAGE = $(OPERATOR_IMAGE_NAME):$(OPERATOR_VERSION)

BUNDLE_IMAGE_NAME = $(OPERATOR_IMAGE_NAME)-bundle
BUNDLE_IMAGE = $(BUNDLE_IMAGE_NAME):$(OPERATOR_VERSION)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false" # TODO

NAMESPACE ?= "apicurio-registry-operator-namespace"

CLIENT ?= kubectl

OPERATOR_SDK = $(shell which operator-sdk)

########## Env

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

########## Tools

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

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: install-controller-gen
install-controller-gen: ## Install controller-gen@v0.4.1
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: install-kustomize
install-kustomize: ## Install kustomize@v3.8.7
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

########## Targets

.PHONY: generate
generate: install-controller-gen ## Generate code
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run 'go fmt' against code
	go fmt ./...

.PHONY: vet
vet: ## Run 'go vet' against code
	go vet ./...

.PHONY: manager
manager: generate fmt vet ## Build manager binary
	go build -o bin/manager main.go

.PHONY: manifests
manifests: install-controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=apicurio-registry-operator-role paths="./..." output:crd:artifacts:config=config/crd/resources output:rbac:artifacts:config=config/rbac/resources

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
.PHONY: test
test: generate fmt vet manifests ## Run tests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

.PHONY: deploy
deploy: manifests install-kustomize ## TODO
	cd config/manager && $(KUSTOMIZE) edit set image REGISTRY_OPERATOR_IMAGE=${OPERATOR_IMAGE}
	cd config/default && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/default | $(CLIENT) apply -f - -n ${NAMESPACE}

.PHONY: undeploy
undeploy: ## TODO
	$(KUSTOMIZE) build config/build-namespaced | $(CLIENT) delete -f -

.PHONY: docker-build
docker-build: test ## Build Operator image
	docker build -t ${OPERATOR_IMAGE} .

.PHONY: docker-push
docker-push: ## Push Operator image
ifeq ($(LATEST),true)
	docker tag $(OPERATOR_IMAGE) $(OPERATOR_IMAGE_NAME):latest$(DASH_VERSION)
	docker push $(OPERATOR_IMAGE_NAME):latest$(DASH_VERSION)
endif
	docker push ${OPERATOR_IMAGE}

.PHONY: build
build: manager docker-build ## Build Operator image

.PHONY: bundle
bundle: manifests install-kustomize ## Generate bundle manifests and metadata
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image REGISTRY_OPERATOR_IMAGE=$(OPERATOR_IMAGE)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(OPERATOR_VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: bundle-build
bundle-build: bundle ## Build bundle image
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMAGE) .

.PHONY: bundle-push
bundle-push: ## Push bundle image
ifeq ($(LATEST),true)
	docker tag $(BUNDLE_IMAGE) $(BUNDLE_IMAGE_NAME):latest$(DASH_VERSION)
	docker push $(BUNDLE_IMAGE_NAME):latest$(DASH_VERSION)
endif
	docker push $(BUNDLE_IMAGE)

DATE=$(shell date -Idate)
.PHONY: packagemanifests
packagemanifests: install-kustomize manifests ## Generate package manifests
	echo "⚠️ Warning: This command requires 'yq' version 4.x"
	which yq
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image REGISTRY_OPERATOR_IMAGE=$(OPERATOR_IMAGE)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate packagemanifests -q --version $(PACKAGE_VERSION) $(PACKAGE_MANIFESTS_OPTS)
	yq e ".metadata.annotations.createdAt = \"$(DATE)\"" -i \
		"packagemanifests/$(PACKAGE_VERSION)/apicurio-registry-operator.clusterserviceversion.yaml"
	yq e ".metadata.annotations.containerImage = \"$(OPERATOR_IMAGE)\"" -i \
		"packagemanifests/$(PACKAGE_VERSION)/apicurio-registry-operator.clusterserviceversion.yaml"

.PHONY: docs
docs: ## Build documentation
	cd ./docs && antora local-test-playbook.yml
	echo "file:$(shell pwd)/docs/target/dist/"

.PHONY: dist
dist: install-kustomize docs ## Generate distribution bundle
	mkdir -p dist
	cp -rt ./dist ./dist-base/*
	cp -t ./dist ./LICENSE
	# Examples
	cp -t ./dist/examples ./config/examples/resources/*
	cp -rt ./dist/examples ./docs/modules/ROOT/examples/*
	#cp -t ./dist/examples/keycloak ./docs/modules/ROOT/examples/keycloak/*
	# Docs
	cp -rt ./dist ./docs/target/dist && mv ./dist/dist ./dist/docs
	# Install
	cd config/manager && $(KUSTOMIZE) edit set image REGISTRY_OPERATOR_IMAGE=$(OPERATOR_IMAGE)
	$(KUSTOMIZE) build config/default/ > ./dist/install.yaml
	$(KUSTOMIZE) build config/default/ > ./docs/resources/install.yaml # Deprecated!
	# Archive
	tar -zcf apicurio-registry-operator-$(PACKAGE_VERSION).tar.gz -C ./dist .

.PHONY: clean
clean: ## Remove temporary and generated files
	rm apicurio-registry-operator-$(PACKAGE_VERSION).tar.gz cover.out || true
	rm -r bin build bundle dist docs/target testbin || true
