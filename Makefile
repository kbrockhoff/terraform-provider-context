# Terraform Provider Context - Makefile

# Variables
PROVIDER_NAME = terraform-provider-context
BINARY_NAME = terraform-provider-brockhoff
VERSION ?= dev
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
TERRAFORM_PLUGINS_DIR = ~/.terraform.d/plugins
REGISTRY_NAMESPACE = registry.terraform.io/kbrockhoff/context
LOCAL_PROVIDER_DIR = $(TERRAFORM_PLUGINS_DIR)/$(REGISTRY_NAMESPACE)/$(VERSION)/$(GOOS)_$(GOARCH)

# Determine Terraform CLI config file path
# Use CI-specific path if running in GitHub Actions, otherwise use local path
ifdef GITHUB_ACTIONS
	TF_CLI_CONFIG_FILE = $(GITHUB_WORKSPACE)/.terraformrc
else
	TF_CLI_CONFIG_FILE = $(PWD)/.terraformrc
endif

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Terraform Provider Context - Development Tasks"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"}; /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Build targets
.PHONY: build
build: ## Build the provider binary
	@echo "Building provider..."
	go build -o bin/$(BINARY_NAME) -ldflags="-X main.version=$(VERSION)"
	@echo "✓ Provider built: bin/$(BINARY_NAME)"

.PHONY: build-all
build-all: ## Build provider for all supported platforms
	@echo "Building provider for all platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)_linux_amd64 -ldflags="-X main.version=$(VERSION)"
	GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)_linux_arm64 -ldflags="-X main.version=$(VERSION)"
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)_darwin_amd64 -ldflags="-X main.version=$(VERSION)"
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)_darwin_arm64 -ldflags="-X main.version=$(VERSION)"
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)_windows_amd64.exe -ldflags="-X main.version=$(VERSION)"
	@echo "✓ Cross-platform binaries built in bin/"

.PHONY: install
install: build ## Install provider locally for development
	@echo "Installing provider locally..."
	@mkdir -p $(LOCAL_PROVIDER_DIR)
	@cp bin/$(BINARY_NAME) $(LOCAL_PROVIDER_DIR)/terraform-provider-context_v$(VERSION)
	@echo "✓ Provider installed to: $(LOCAL_PROVIDER_DIR)"
	@echo "Creating .terraformrc from template..."
	@sed "s|PROVIDER_PATH|$(shell echo $(LOCAL_PROVIDER_DIR) | sed 's|~|$(HOME)|g')|g" .terraformrc.tmpl > $(TF_CLI_CONFIG_FILE)
	@echo "✓ Created $(TF_CLI_CONFIG_FILE)"
	@echo "✓ You can now use the provider in your Terraform configurations"

# Testing targets
.PHONY: test
test: ## Run unit tests
	@echo "Running unit tests..."
	go test -v ./internal/...
	@echo "✓ Unit tests passed"

.PHONY: test-race
test-race: ## Run unit tests with race detection
	@echo "Running unit tests with race detection..."
	go test -race -v ./internal/...
	@echo "✓ Unit tests with race detection passed"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./internal/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "✓ Coverage report generated: coverage/coverage.html"
	go tool cover -func=coverage/coverage.out | grep total:

.PHONY: test-examples
test-examples: install ## Test example configurations
	@echo "Testing example configurations..."
	@for dir in examples/*/; do \
		if [ -d "$$dir" ]; then \
			case "$$(basename $$dir)" in \
				data-sources) \
					continue ;; \
			esac; \
			echo "Testing $$dir..."; \
			cd "$$dir" && \
			TF_CLI_CONFIG_FILE="$(TF_CLI_CONFIG_FILE)" terraform init -upgrade && \
			TF_CLI_CONFIG_FILE="$(TF_CLI_CONFIG_FILE)" terraform validate && \
			TF_CLI_CONFIG_FILE="$(TF_CLI_CONFIG_FILE)" terraform plan; \
			if [ $$? -ne 0 ]; then \
				echo "❌ Example $$dir failed"; \
				exit 1; \
			fi; \
			cd ../..; \
		fi; \
	done
	@echo "✓ All examples validated successfully"

.PHONY: acceptance-test
acceptance-test: install ## Run Terraform acceptance tests
	@echo "Running acceptance tests..."
	TF_ACC=1 go test -v ./internal/provider/... -timeout 120m
	@echo "✓ Acceptance tests passed"

# Code quality targets
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "✓ Code formatted"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ go vet passed"

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint)
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "✓ Linting passed"; \
	else \
		echo "❌ golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

.PHONY: check
check: fmt vet test ## Run all code quality checks
	@echo "✓ All checks passed"

# Documentation targets
.PHONY: docs-generate
docs-generate: ## Generate provider documentation
	@echo "Generating provider documentation..."
	@if command -v tfplugindocs >/dev/null 2>&1; then \
		tfplugindocs generate --provider-name=brockhoff; \
		echo "✓ Documentation generated"; \
	else \
		echo "❌ tfplugindocs not found. Install it with: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"; \
		exit 1; \
	fi

.PHONY: docs-validate
docs-validate: ## Validate provider documentation
	@echo "Validating provider documentation..."
	@if command -v tfplugindocs >/dev/null 2>&1; then \
		tfplugindocs validate --provider-name=brockhoff; \
		echo "✓ Documentation validated"; \
	else \
		echo "❌ tfplugindocs not found. Install it with: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"; \
		exit 1; \
	fi

# Development targets
.PHONY: dev-setup
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	go mod download
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v tfplugindocs >/dev/null 2>&1; then \
		echo "Installing tfplugindocs..."; \
		go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest; \
	fi
	@echo "✓ Development environment ready"

.PHONY: mod-tidy
mod-tidy: ## Clean up go.mod and go.sum
	@echo "Tidying Go modules..."
	go mod tidy
	@echo "✓ Go modules tidied"

.PHONY: mod-verify
mod-verify: ## Verify module dependencies
	@echo "Verifying module dependencies..."
	go mod verify
	@echo "✓ Module dependencies verified"

.PHONY: dev-reset
dev-reset: clean ## Reset development environment
	@echo "Resetting development environment..."
	@rm -rf ~/.terraform.d/plugins/$(REGISTRY_NAMESPACE)
	go clean -modcache
	@echo "✓ Development environment reset"

# Release targets
.PHONY: release-snapshot
release-snapshot: ## Create a snapshot release (requires goreleaser)
	@echo "Creating snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
		echo "✓ Snapshot release created"; \
	else \
		echo "❌ goreleaser not found. Install it from: https://goreleaser.com/install/"; \
		exit 1; \
	fi

.PHONY: release-dry-run
release-dry-run: ## Dry run release process
	@echo "Running release dry-run..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --skip=publish --clean; \
		echo "✓ Release dry-run completed"; \
	else \
		echo "❌ goreleaser not found. Install it from: https://goreleaser.com/install/"; \
		exit 1; \
	fi

# Utility targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -rf coverage/
	@echo "✓ Build artifacts cleaned"

.PHONY: version
version: ## Show current version information
	@echo "Version: $(VERSION)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
	@echo "Registry: $(REGISTRY_NAMESPACE)"

.PHONY: debug-install
debug-install: ## Install provider with debug information
	@echo "Installing provider with debug support..."
	go build -gcflags="all=-N -l" -o bin/$(BINARY_NAME)_debug -ldflags="-X main.version=$(VERSION)-debug"
	@mkdir -p $(LOCAL_PROVIDER_DIR)
	@cp bin/$(BINARY_NAME)_debug $(LOCAL_PROVIDER_DIR)/terraform-provider-context_v$(VERSION)
	@echo "✓ Debug provider installed"
	@echo "✓ Use 'dlv attach' to debug the provider process"

.PHONY: terraform-init
terraform-init: install ## Initialize Terraform in examples directory
	@echo "Initializing Terraform examples..."
	@cd examples && TF_CLI_CONFIG_FILE="$(TF_CLI_CONFIG_FILE)" terraform init -upgrade
	@echo "✓ Terraform initialized in examples/"

.PHONY: terraform-plan
terraform-plan: terraform-init ## Run terraform plan on examples
	@echo "Running terraform plan on examples..."
	@cd examples && TF_CLI_CONFIG_FILE="$(TF_CLI_CONFIG_FILE)" terraform plan -out=tfplan
	@echo "✓ Terraform plan completed"

.PHONY: terraform-apply
terraform-apply: terraform-plan ## Apply terraform configuration (examples)
	@echo "Applying terraform configuration..."
	@cd examples && TF_CLI_CONFIG_FILE="$(TF_CLI_CONFIG_FILE)" terraform apply tfplan
	@echo "✓ Terraform apply completed"

.PHONY: terraform-destroy
terraform-destroy: ## Destroy terraform resources (examples)
	@echo "Destroying terraform resources..."
	@cd examples && TF_CLI_CONFIG_FILE="$(TF_CLI_CONFIG_FILE)" terraform destroy -auto-approve
	@echo "✓ Terraform resources destroyed"

# CI/CD targets
.PHONY: ci
ci: check test ## Run CI pipeline tasks
	@echo "✓ CI pipeline completed successfully"

.PHONY: pre-commit
pre-commit: fmt vet test ## Run pre-commit checks
	@echo "✓ Pre-commit checks passed"

# Info targets
.PHONY: status
status: ## Show development environment status
	@echo "=== Development Environment Status ==="
	@echo "Go version: $$(go version)"
	@echo "Provider version: $(VERSION)"
	@echo "Build target: $(GOOS)/$(GOARCH)"
	@echo ""
	@echo "=== Tools Status ==="
	@printf "golangci-lint: "; if command -v golangci-lint >/dev/null 2>&1; then echo "✓ installed"; else echo "❌ not installed"; fi
	@printf "tfplugindocs: "; if command -v tfplugindocs >/dev/null 2>&1; then echo "✓ installed"; else echo "❌ not installed"; fi
	@printf "goreleaser:   "; if command -v goreleaser >/dev/null 2>&1; then echo "✓ installed"; else echo "❌ not installed"; fi
	@printf "terraform:    "; if command -v terraform >/dev/null 2>&1; then echo "✓ installed ($$(terraform version -json | jq -r .terraform_version))"; else echo "❌ not installed"; fi
	@echo ""
	@echo "=== Provider Status ==="
	@if [ -f "$(LOCAL_PROVIDER_DIR)/$(BINARY_NAME)_v$(VERSION)" ]; then \
		echo "Provider: ✓ installed locally"; \
		echo "Location: $(LOCAL_PROVIDER_DIR)"; \
	else \
		echo "Provider: ❌ not installed locally (run 'make install')"; \
	fi

.PHONY: install-tools
install-tools: ## Install all development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	@echo "✓ Development tools installed"

# Phony target to ensure all targets are run regardless of file existence
.PHONY: all
all: clean dev-setup check build test install ## Run complete development workflow