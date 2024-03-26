##@ Build Dependencies Binary

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)


## Tools

# Set the Operator SDK version to use. By default, what is installed on the system is used.
# This is useful for CI or a project to utilize a specific version of the operator-sdk toolkit.
OPERATOR_SDK_VERSION ?= v1.31.0
.PHONY: operator-sdk
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
operator-sdk: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(OPERATOR_SDK)))
ifeq (, $(shell which operator-sdk 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPERATOR_SDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPERATOR_SDK) ;\
	}
else
OPERATOR_SDK = $(shell which operator-sdk)
endif
endif


KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v5.1.1
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }


CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.13.0
.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)


ENV_TEST ?= $(LOCALBIN)/setup-envtest
.PHONY: kubebuilder_assets, env-test
env-test: $(ENV_TEST) ## Download env-test-setup locally if necessary.
$(ENV_TEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20230216140739-c98506dc3b8e

# ENV_TEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by env-test binary.
ENV_TEST_K8S_VERSION = 1.26.0

OS=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)
KUBEBUILDER_TOOL_ASSETS ?= $(LOCALBIN)/k8s/$(ENV_TEST_K8S_VERSION)-$(OS)-$(ARCH)


.PHONY: kubebuilder_assets
kubebuilder_assets:           ## Download kubebuilder tools locally if necessary.
ifeq (, $(wildcard $(KUBEBUILDER_TOOL_ASSETS)))
	@{ \
	set -e ;\
	mkdir -p $(KUBEBUILDER_TOOL_ASSETS);\
	}
else
	@{ \
	set -e ;\
	rm -rf $(KUBEBUILDER_TOOL_ASSETS);\
	}
endif
	wget -O kubebuilder-tools.tar.gz $(ARTIFACTS_SERVER)/kubebuilder-tools-$(ENV_TEST_K8S_VERSION)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz;\
	tar -xzf kubebuilder-tools.tar.gz -C $(LOCALBIN);\
	mv $(LOCALBIN)/kubebuilder/bin/* $(KUBEBUILDER_TOOL_ASSETS)/ ;\
	rm -rf kubebuilder-tools.tar.gz $(LOCALBIN)/kubebuilder;\
KUBEBUILDER_ASSETS=$(KUBEBUILDER_TOOL_ASSETS)
$(info KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS))

# golang-ci
GOLANG_CI_LINT_VERSION ?= v1.50.0
.PHONY: golang-ci
golang-ci:  ## Download Golang ci-lint locally if necessary.
ifneq ($(shell which golangci-lint),)
	@$(OK) "golangci-lint is already installed"
GOLANG_CI_LINT=$(shell which golangci-lint)
else ifeq (, $(shell which $(GOBIN)/golangci-lint))
	@{ \
	set -e ;\
	echo 'installing golangci-lint-$(GOLANG_CI_LINT_VERSION)' ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCILINT_VERSION) ;\
	echo 'Successfully installed' ;\
	}
GOLANG_CI_LINT=$(GOBIN)/golangci-lint
else
	@$(OK) "golangci-lint is already installed"
GOLANG_CI_LINT=$(GOBIN)/golangci-lint
endif

.PHONY: static-check
static-check: ## Download staticcheck locally if necessary.
ifeq (, $(shell which staticcheck))
	@{ \
	set -e ;\
	echo 'installing honnef.co/go/tools/cmd/staticcheck ' ;\
	go install honnef.co/go/tools/cmd/staticcheck@2022.1 ;\
	}
STATIC_CHECK=$(GOBIN)/staticcheck
else
STATIC_CHECK=$(shell which staticcheck)
endif

.PHONY: goimports
goimports: ## Download goimports locally if necessary.
ifeq (, $(shell which goimports))
	@{ \
	set -e ;\
	go install golang.org/x/tools/cmd/goimports@latest ;\
	}
GOIMPORTS=$(GOBIN)/goimports
else
GOIMPORTS=$(shell which goimports)
endif

HELM_VERSION ?= helm-v3.6.0-linux-amd64.tar.gz
.PHONY: helm
helm: ## Download helm cli locally if necessary.
ifeq (, $(shell which helm))
	@{ \
	set -e ;\
	echo 'installing $(HELM_VERSION)' ;\
	wget $(ARTIFACTS_SERVER)/$(HELM_VERSION) ;\
	tar -zxvf $(HELM_VERSION) ;\
	mv linux-amd64/helm /bin/helm ;\
	rm -f $(HELM_VERSION) ;\
	rm -rf linux-amd64 ;\
	echo 'Successfully installed' ;\
    }
else
	@$(OK) Helm CLI is already installed
HELMBIN=$(shell which helm)
endif


.PHONY: helm-doc
helm-doc: ## Install helm-doc locally if necessary.
ifeq (, $(shell which readme-generator))
	@{ \
	set -e ;\
	echo 'installing readme-generator-for-helm' ;\
	npm install -g @bitnami/readme-generator-for-helm ;\
	}
else
	@$(OK) readme-generator-for-helm is already installed
HELM_DOC=$(shell which readme-generator)
endif

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.23.0/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif