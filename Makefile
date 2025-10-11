.PHONY: build test lint fmt clean install uninstall run help

# Binary name
BINARY=karabiner-monitor

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOLINT=golangci-lint

help: ## Display this help message
	@echo "Karabiner Monitor - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""

build: ## Build the binary
	@echo "Building $(BINARY)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY) ./cmd/$(BINARY)
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY)"

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v -race ./...
	@echo "✓ Tests passed"

lint: ## Run linter
	@echo "Running linter..."
	$(GOLINT) run ./...
	@echo "✓ Lint passed"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "✓ Code formatted"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BUILD_DIR)/$(BINARY)
	@echo "✓ Clean complete"

install: build ## Install the service (requires sudo)
	@echo "Installing $(BINARY)..."
	@./scripts/install.sh

uninstall: ## Uninstall the service (requires sudo)
	@echo "Uninstalling $(BINARY)..."
	@./scripts/uninstall.sh

run: build ## Build and run locally (for testing)
	@echo "Running $(BINARY)..."
	@./$(BINARY)

check: test lint ## Run tests and linter
	@echo "✓ All checks passed"

all: clean fmt lint test build ## Run all: clean, format, lint, test, build
	@echo "✓ All tasks complete"
