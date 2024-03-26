##@ Build Image

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
IMAGE_TAG_BASE ?= linktimecloud

# Image URL to use all building/pushing image targets
IMG_CONTROLLER       ?= $(IMAGE_TAG_BASE)/kdp-oam-operator:$(VERSION)
IMG_CLI              ?= $(IMAGE_TAG_BASE)/kdp-oam-bdcctl:$(VERSION)
IMG_API_SERVER ?= $(IMAGE_TAG_BASE)/kdp-oam-apiserver:$(VERSION)
IMG_REGISTRY   ?= ""

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG_CONTROLLER=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG_CONTROLLER=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG_CONTROLLER} -f Dockerfile.cross .
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross

.PHONY: docker-build
docker-build: docker-build-core docker-build-apiserver docker-build-cli  ## Build docker image with the manager.
	@$(OK)

.PHONY: docker-build-core
docker-build-core:
	docker build --build-arg=VERSION=$(VERSION) --build-arg=GITVERSION=$(GIT_COMMIT) -t $(IMG_REGISTRY)/$(IMG_CONTROLLER) .

.PHONY: docker-build-apiserver
docker-build-apiserver:
	docker build --build-arg=VERSION=$(VERSION) --build-arg=GITCOMMIT=$(GIT_COMMIT) -t $(IMG_REGISTRY)/$(IMG_API_SERVER) -f apiserver.Dockerfile .

.PHONY: docker-build-cli
docker-build-cli:
	docker build -t $(IMG_REGISTRY)/$(IMG_CLI) -f ./images/bdcctl/Dockerfile .

.PHONY: docker-push
docker-push: docker-push-core docker-push-apiserver docker-push-cli  ## Push docker image with the manager.
	@$(OK)

.PHONY: docker-push-core
docker-push-core:
	docker push $(IMG_REGISTRY)/$(IMG_CONTROLLER)

.PHONY: docker-push-apiserver
docker-push-apiserver:
	docker push $(IMG_REGISTRY)/$(IMG_API_SERVER)

.PHONY: docker-push-cli
docker-push-cli:
	docker push $(IMG_REGISTRY)/$(IMG_CLI)