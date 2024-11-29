.DEFAULT_GOAL := build

VERSION 			?= latest

GO 						?= go
GO_RUN_TOOLS 	?= $(GO) run -modfile ./tools/go.mod
GO_TEST 			?= $(GO_RUN_TOOLS) gotest.tools/gotestsum --format pkgname
GO_RELEASER 	?= $(GO_RUN_TOOLS) github.com/goreleaser/goreleaser
GO_MOD 				?= $(shell ${GO} list -m)
GO_KUSTOMIZE 	?= $(GO_RUN_TOOLS) sigs.k8s.io/kustomize/kustomize/v5

IMAGE_TAG_BASE 		?= ghcr.io/zeiss/natz-operator/operator
IMG 				?= $(IMAGE_TAG_BASE):$(VERSION)

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: build
build: ## Build the binary file.
	$(GO_RELEASER) build --snapshot --clean

.PHONY: snapshot
snapshot: ## Create a snapshot release
	$(GO_RELEASER) release --clean --snapshot

.PHONY: release
release: ## Create a release
	$(GO_RELEASER) release --clean

.PHONY: up
up: ## Run the operator locally.
	air -c .air.toml

.PHONY: install
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(GO_KUSTOMIZE) build manifests/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(GO_KUSTOMIZE) build manifests/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(GO_KUSTOMIZE) edit set image controller=${IMG}
	$(GO_KUSTOMIZE) build manifests/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(GO_KUSTOMIZE) build manifests/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: minikube-push
minikube-push: ## Push the image to the minikube docker daemon.
	minikube image rm ${IMG}
	minikube image load ${IMG}

.PHONY: generate
generate: ## Generate code.
	$(GO) generate ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	$(GO_RUN_TOOLS) mvdan.cc/gofumpt -w .

.PHONY: vet
vet: ## Run go vet against code.
	$(GO) vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	mkdir -p .test/reports
	$(GO_TEST) --junitfile .test/reports/unit-test.xml -- -race ./... -count=1 -short -cover -coverprofile .test/reports/unit-test-coverage.out

.PHONY: lint
lint: ## Run lint.
	$(GO_RUN_TOOLS) github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout 5m -c .golangci.yml

.PHONY: clean
clean: ## Remove previous build.
	rm -rf .test .dist
	find . -type f -name '*.gen.go' -exec rm {} +
	git checkout go.mod

.PHONY: help
help: ## Display this help screen.
	@grep -E '^[a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
