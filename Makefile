IMG ?= controller:latest
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
BIN_DIR := bin

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

HOME ?= /tmp/kubebuilder-testing
ifeq ($(HOME), /)
HOME = /tmp/kubebuilder-testing
endif

all: build

verify-%:
	make $*
	./hack/verify-diff.sh

verify: fmt lint

# Run tests
test: verify unit

# Build operator binaries
build: operator

operator:
	go build -o bin/cluster-capi-operator cmd/cluster-capi-operator/main.go

unit: ginkgo envtest
	KUBEBUILDER_ASSETS=$(shell $(ENVTEST) --bin-dir=$(shell pwd)/bin use $(ENVTEST_K8S_VERSION) -p path) ./hack/test.sh "./pkg/... ./assets/..." 5m

.PHONY: e2e
e2e: ginkgo
	./hack/test.sh "./e2e/..." 30m

# Run against the configured Kubernetes cluster in ~/.kube/config
run: verify
	go run cmd/cluster-capi-operator/main.go --leader-elect=false --images-json=./hack/sample-images.json

# Run go fmt against code
.PHONY: fmt
fmt: golangci-lint
	( GOLANGCI_LINT_CACHE=$(PROJECT_DIR)/.cache $(GOLANGCI_LINT) run --fix )

# Run go vet against code
.PHONY: vet
vet: lint

.PHONY: golangci-lint lint
lint: golangci-lint
	( GOLANGCI_LINT_CACHE=$(PROJECT_DIR)/.cache $(GOLANGCI_LINT) run )

.PHONY: bin
bin:
	mkdir -p $(BIN_DIR)

# Download golangci-lint locally if necessary
.PHONY: golangci-lint
GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
golangci-lint: bin
	pushd "$(PROJECT_DIR)"/hack/tools && go build -o "../../$(BIN_DIR)" ./vendor/github.com/golangci/golangci-lint/cmd/golangci-lint && popd

.PHONY: envtest
ENVTEST = $(shell pwd)/bin/setup-envtest
envtest: bin
	pushd "$(PROJECT_DIR)"/hack/tools && go build -o "../../$(BIN_DIR)" ./vendor/sigs.k8s.io/controller-runtime/tools/setup-envtest && popd

.PHONY: ginkgo
GINKGO = $(shell pwd)/bin/GINKGO
ginkgo: bin
	pushd "$(PROJECT_DIR)"/hack/tools && go build -o "../../$(BIN_DIR)" ./vendor/github.com/onsi/ginkgo/v2/ginkgo && popd

.PHONY: kustomize
KUSTOMIZE = $(shell pwd)/bin/KUSTOMIZE
kustomize: bin
	pushd "$(PROJECT_DIR)"/hack/tools && go build -o "../../$(BIN_DIR)" ./vendor/sigs.k8s.io/kustomize/kustomize/v3 && popd

.PHONY: assets
assets:
	./hack/assets.sh

# Run go mod
.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

# Build the docker image
.PHONY: image
image:
	docker build -t ${IMG} .

# Push the docker image
.PHONY: push
push:
	docker push ${IMG}

aws-cluster:
	./hack/clusters/create-aws.sh

azure-cluster:
	./hack/clusters/create-azure.sh

gcp-cluster:
	./hack/clusters/create-gcp.sh
