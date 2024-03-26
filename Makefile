include makefiles/const.mk
include makefiles/build-image.mk
include makefiles/build-swagger.mk
include makefiles/build-bundle.mk
include makefiles/build-catalog.mk
include makefiles/build-helm-package.mk
include makefiles/dependency.mk

# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.0.1

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
cli: ## Generate bdcctl cli.
	go build -o bdcctl reference/cmd/cli/main.go

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./api/..." output:crd:artifacts:config=config/crd/bases

.PHONY: cp
cp: ## Copy CRD files to helm chart.
	cp -r config/crd/bases/*.yaml charts/kdp-oam-operator/crds

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...


.PHONY: test
test: manifests generate cp fmt vet env-test ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENV_TEST) use $(ENV_TEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $(shell go list ./pkg/...) -coverprofile cover.out && go tool cover -func=cover.out
	$(ENV_TEST) cleanup $(ENV_TEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path

.PHONY: clean
clean: ## clean up env-test binary.
	if [[ -d $(LOCALBIN)/k8s ]]; then \
        rm -rf $(LOCALBIN)/k8s; \
    fi


##@ Deployment

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/bdc/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run cmd/bdc/main.go

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -


.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG_CONTROLLER}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: apply
apply: manifests kustomize ## Apply samples into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/samples | kubectl apply -f -

.PHONY: delete
delete: manifests kustomize ## Delete samples from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/samples | kubectl delete -f -

.PHONY: run-apiserver
run-apiserver: fmt vet ## Run a api server from your host.
	go run cmd/apiserver/main.go --kube-api-qps=300 --kube-api-burst=900
