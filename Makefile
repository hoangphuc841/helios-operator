# =============================================================================
# 🚀 Helios Operator Makefile - IDP Platform
# =============================================================================
# Cloud-Native Internal Developer Platform Operator
# Zero-Configuration GitOps Deployment for Kubernetes Applications
# =============================================================================

# =============================================================================
# 🎯 Core Configuration
# =============================================================================

# Project Information
PROJECT_NAME := helios-operator
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Container Configuration
REGISTRY ?= helios.dev
IMAGE_NAME ?= $(PROJECT_NAME)
IMG ?= $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
LATEST_IMG ?= $(REGISTRY)/$(IMAGE_NAME):latest

# Go Configuration
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

# Kubernetes Configuration
KIND_CLUSTER ?= helios-test
NAMESPACE ?= helios-system

# Development Configuration
DEV_MODE ?= false

# Directories
BIN_DIR := $(shell pwd)/bin

# =============================================================================
# 🛠️ Tool Configuration
# =============================================================================

# Essential Tools
KUBECTL ?= kubectl
KIND ?= kind
HELM ?= helm
KUSTOMIZE ?= $(BIN_DIR)/kustomize
CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
GOLANGCI_LINT ?= $(BIN_DIR)/golangci-lint

# Tool Versions
KUSTOMIZE_VERSION ?= v5.7.1
CONTROLLER_TOOLS_VERSION ?= v0.19.0
GOLANGCI_LINT_VERSION ?= v2.5.0
ENVTEST_K8S_VERSION ?= 1.34.1
ENVTEST ?= $(BIN_DIR)/setup-envtest

# =============================================================================
# 🎯 Default Target
# =============================================================================

.PHONY: all
all: help

# =============================================================================
# 📚 Help System
# =============================================================================

.PHONY: help
help: ## 📖 Display help for Helios IDP Operator
	@echo "============================================================================="
	@echo "🚀 Helios Operator - Internal Developer Platform"
	@echo "============================================================================="
	@echo "Zero-Configuration GitOps Deployment for Kubernetes Applications"
	@echo "Version: $(VERSION) | Commit: $(COMMIT)"
	@echo "============================================================================="
	@echo ""
	@echo "Core Commands:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Core/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Development:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Development/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "GitOps & Deploy:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- GitOps/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Testing:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Test/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Security:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Security/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Build:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Build/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Documentation:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Documentation/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Container:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Container/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Helm:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Helm/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Tools:"
	@awk 'BEGIN {FS = ":.*?##"} /^[a-zA-Z0-9_-]+:.*?##.*- Tools/ { printf "  \033[36m%-25s\033[0m  \033[37m%s\033[0m\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Quick Start:"
	@printf "  \033[36mmake dev-setup\033[0m         # Setup development environment\n"
	@printf "  \033[36mmake deploy-local\033[0m      # Deploy to local cluster\n"
	@printf "  \033[36mmake test-e2e\033[0m          # Run end-to-end tests\n"
	@printf "  \033[36mmake validate-security\033[0m # Validate security compliance\n"
	@printf "  \033[36mmake helm-package\033[0m      # Package Helm chart\n"
	@echo ""

# =============================================================================
# 🎯 Core Targets
# =============================================================================

.PHONY: build
build: manifests generate ## Build the Helios operator binary - Core
	@echo "Building Helios operator..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "-w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)" \
		-o $(BIN_DIR)/manager cmd/main.go
	@echo "Build complete: $(BIN_DIR)/manager"

.PHONY: run
run: manifests generate ## Run the operator locally - Core
	@echo "Running Helios operator..."
	@go run ./cmd/main.go

.PHONY: manifests
manifests: controller-gen ## Generate CRDs and RBAC manifests - Core
	@echo "Generating manifests..."
	@$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths=./api/... output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate DeepCopy code - Core
	@echo "Generating code..."
	@$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/v1"

# =============================================================================
# 🔧 Development Targets
# =============================================================================

.PHONY: dev-setup
dev-setup: tools ## Setup development environment for Helios IDP - Development
	@echo "Setting up Helios development environment..."
	@$(MAKE) clean
	@$(MAKE) generate manifests
	@$(MAKE) build
	@echo "Development environment ready!"

.PHONY: fmt
fmt: ## Format Go code - Development
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet - Development
	@echo "Vetting code..."
	@go vet ./...

.PHONY: lint
lint: golangci-lint ## Run linter - Development
	@echo "Running linter..."
	@$(GOLANGCI_LINT) run --timeout=5m

.PHONY: test
test: manifests generate setup-envtest-bins ## Run unit tests - Development
	@echo "Running tests..."
	@go test ./... -race -v

# =============================================================================
# 🧪 Testing Targets
# =============================================================================

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests for GitOps workflow - Test
	@echo "Running e2e tests for Helios GitOps workflow..."
	@go test ./test/e2e/ -v -timeout=30m

.PHONY: test-coverage
test-coverage: ## Run tests with coverage - Test
	@echo "Running tests with coverage..."
	@go test ./... -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: validate-security
validate-security: ## Validate Pod Security Standards compliance - Security
	@echo "🔒 Validating Pod Security Standards compliance..."
	@./scripts/validate-security.sh

.PHONY: verify
verify: fmt vet lint test ## Run all verification checks - Development
	@echo "All verification checks passed!"

# =============================================================================
# 📦 Build Targets
# =============================================================================

.PHONY: build-all
build-all: ## Build for all platforms - Build
	@echo "Building for all platforms..."
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "Building for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -o bin/$(PROJECT_NAME)-$$os-$$arch ./cmd/main.go; \
		done \
	done
	@echo "Multi-platform build complete!"

# =============================================================================
# 📚 Documentation Targets
# =============================================================================

.PHONY: docs
docs: api-docs ## Generate all documentation - Documentation
	@echo "Documentation generation complete!"

.PHONY: api-docs
api-docs: controller-gen ## Generate API reference documentation - Documentation
	@echo "Generating API documentation..."
	@mkdir -p docs/reference
	@$(CONTROLLER_GEN) crd:crdVersions=v1 paths=./api/... output:crd:dir=./docs/reference/
	@echo "API documentation generated in docs/reference/"

# =============================================================================
# 🐳 Container Targets
# =============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image for Helios operator - Container
	@echo "Building Docker image: $(IMG)"
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--tag $(IMG) \
		--tag $(LATEST_IMG) \
		.

# =============================================================================
# 🚀 GitOps & Deploy Targets
# =============================================================================

.PHONY: install
install: manifests kustomize ## Install Helios CRDs - GitOps
	@echo "Installing Helios CRDs..."
	@$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy Helios operator - GitOps
	@echo "Deploying Helios operator..."
	@cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	@$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: deploy-local
deploy-local: ## Deploy to local Kind cluster for GitOps testing - GitOps
	@echo "Deploying to local Kind cluster..."
	@$(MAKE) docker-build
	@$(KIND) load docker-image $(IMG) --name $(KIND_CLUSTER)
	@$(MAKE) deploy IMG=$(IMG)

.PHONY: undeploy
undeploy: kustomize ## Undeploy Helios operator - GitOps
	@echo "Undeploying Helios operator..."
	@$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=true -f -

# =============================================================================
# 📊 Helm Targets
# =============================================================================

.PHONY: helm-package
helm-package: ## Package Helm chart for IDP distribution - Helm
	@echo "Packaging Helios Helm chart..."
	@$(HELM) package helm/$(PROJECT_NAME) --destination dist/

.PHONY: helm-lint
helm-lint: ## Lint Helm chart - Helm
	@echo "Linting Helm chart..."
	@$(HELM) lint helm/$(PROJECT_NAME)

.PHONY: helm-deploy
helm-deploy: ## Install Helm chart - Helm
	@echo "Installing Helios via Helm..."
	@$(HELM) install $(PROJECT_NAME) helm/$(PROJECT_NAME) -n $(NAMESPACE) --create-namespace

.PHONY: status
status: ## Check Helios operator status - Helm
	@echo "Checking Helios operator status..."
	@$(KUBECTL) get pods -n $(NAMESPACE) -l app.kubernetes.io/name=$(PROJECT_NAME)
	@$(KUBECTL) get heliosapp -A

# =============================================================================
# 🧹 Cleanup Targets
# =============================================================================

.PHONY: clean
clean: ## Clean build artifacts - Tools
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# =============================================================================
# 🛠️ Tool Installation
# =============================================================================

.PHONY: tools
tools: kustomize controller-gen golangci-lint ## Install essential tools for Helios IDP - Tools

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Install kustomize - Tools
$(KUSTOMIZE): $(BIN_DIR)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Install controller-gen - Tools
$(CONTROLLER_GEN): $(BIN_DIR)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Install golangci-lint - Tools
$(GOLANGCI_LINT): $(BIN_DIR)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Install setup-envtest - Tools
$(ENVTEST): $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: setup-envtest-bins
setup-envtest-bins: envtest ## Download envtest K8s binaries - Tools
	@echo "📦 Installing envtest Kubernetes binaries..."
	@ENVTEST_ASSETS_DIR=$$($(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(shell pwd)/bin/k8s -p path) && \
	echo "Envtest binaries installed at: $$ENVTEST_ASSETS_DIR"

# =============================================================================
# 🔧 Utility Functions
# =============================================================================

# go-install-tool installs Go tools if they don't exist
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "📦 Installing $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(BIN_DIR) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

# =============================================================================
# 📁 Directory Creation
# =============================================================================

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)
