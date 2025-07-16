.PHONY: help build clean test lint fmt vet deps install uninstall run dev release packages docker

# Application name and version
APP_NAME := serverhealth
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS := -ldflags="-w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"
BUILD_FLAGS := $(LDFLAGS) -trimpath

# Directories
BUILD_DIR := build
DIST_DIR := dist
BIN_DIR := bin

# Colors for terminal output
BLUE := \033[34m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

# Default target
help: ## Show this help message
	@echo "$(BLUE)Server Monitor Build System$(RESET)"
	@echo "$(BLUE)=============================$(RESET)"
	@echo ""
	@echo "$(GREEN)Available targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(GREEN)Examples:$(RESET)"
	@echo "  make build          # Build for current platform"
	@echo "  make release        # Build for all platforms"
	@echo "  make packages       # Create distribution packages"
	@echo "  make clean          # Clean build artifacts"
	@echo ""

# Development targets
dev: deps fmt lint test build ## Complete development workflow
	@echo "$(GREEN)✓ Development build complete$(RESET)"

run: build ## Build and run the application
	@echo "$(BLUE)Running $(APP_NAME)...$(RESET)"
	@./$(BIN_DIR)/$(APP_NAME) --help

install-local: build ## Install to local bin directory
	@echo "$(BLUE)Installing $(APP_NAME) to ~/.local/bin/...$(RESET)"
	@mkdir -p ~/.local/bin
	@cp $(BIN_DIR)/$(APP_NAME) ~/.local/bin/
	@echo "$(GREEN)✓ Installed to ~/.local/bin/$(APP_NAME)$(RESET)"

# Build targets
build: deps ## Build for current platform
	@echo "$(BLUE)Building $(APP_NAME) v$(VERSION)...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(APP_NAME) .
	@echo "$(GREEN)✓ Build complete: $(BIN_DIR)/$(APP_NAME)$(RESET)"

build-linux: deps ## Build for Linux
	@echo "$(BLUE)Building for Linux...$(RESET)"
	@mkdir -p $(BUILD_DIR)/linux
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME)-amd64 .
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME)-arm64 .
	@GOOS=linux GOARCH=386 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME)-386 .
	@echo "$(GREEN)✓ Linux builds complete$(RESET)"

build-darwin: deps ## Build for macOS
	@echo "$(BLUE)Building for macOS...$(RESET)"
	@mkdir -p $(BUILD_DIR)/darwin
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/darwin/$(APP_NAME)-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/darwin/$(APP_NAME)-arm64 .
	@echo "$(GREEN)✓ macOS builds complete$(RESET)"

build-windows: deps ## Build for Windows
	@echo "$(BLUE)Building for Windows...$(RESET)"
	@mkdir -p $(BUILD_DIR)/windows
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/windows/$(APP_NAME)-amd64.exe .
	@GOOS=windows GOARCH=386 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/windows/$(APP_NAME)-386.exe .
	@echo "$(GREEN)✓ Windows builds complete$(RESET)"

build-freebsd: deps ## Build for FreeBSD
	@echo "$(BLUE)Building for FreeBSD...$(RESET)"
	@mkdir -p $(BUILD_DIR)/freebsd
	@GOOS=freebsd GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/freebsd/$(APP_NAME)-amd64 .
	@echo "$(GREEN)✓ FreeBSD builds complete$(RESET)"

release: clean build-linux build-darwin build-windows build-freebsd ## Build for all platforms
	@echo "$(BLUE)Creating release packages...$(RESET)"
	@chmod +x build.sh
	@VERSION=$(VERSION) ./build.sh
	@echo "$(GREEN)✓ Release build complete$(RESET)"

# Package targets
packages: release ## Create distribution packages (.deb, .rpm)
	@echo "$(BLUE)Creating distribution packages...$(RESET)"
	@if command -v dpkg-deb >/dev/null 2>&1; then \
		chmod +x create_deb.sh; \
		VERSION=$(VERSION) ./create_deb.sh; \
	else \
		echo "$(YELLOW)⚠ dpkg-deb not found, skipping .deb creation$(RESET)"; \
	fi
	@if command -v rpmbuild >/dev/null 2>&1; then \
		chmod +x create_rpm.sh; \
		VERSION=$(VERSION) ./create_rpm.sh; \
	else \
		echo "$(YELLOW)⚠ rpmbuild not found, skipping .rpm creation$(RESET)"; \
	fi
	@echo "$(GREEN)✓ Package creation complete$(RESET)"

# Docker targets
docker: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(RESET)"
	@docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@echo "$(GREEN)✓ Docker image built: $(APP_NAME):$(VERSION)$(RESET)"

docker-run: docker ## Build and run Docker container
	@echo "$(BLUE)Running Docker container...$(RESET)"
	@docker run -it --rm $(APP_NAME):latest

# Quality assurance targets
test: ## Run tests
	@echo "$(BLUE)Running tests...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests passed$(RESET)"

test-coverage: test ## Run tests with coverage report
	@echo "$(BLUE)Generating coverage report...$(RESET)"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(RESET)"

lint: ## Run linter
	@echo "$(BLUE)Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not found, install from https://golangci-lint.run/$(RESET)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✓ Linting complete$(RESET)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet complete$(RESET)"

# Dependency management
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

deps-update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

deps-vendor: ## Vendor dependencies
	@echo "$(BLUE)Vendoring dependencies...$(RESET)"
	@go mod vendor
	@echo "$(GREEN)✓ Dependencies vendored$(RESET)"

# Installation targets
install: build ## Install system-wide (requires sudo)
	@echo "$(BLUE)Installing $(APP_NAME) system-wide...$(RESET)"
	@sudo cp $(BIN_DIR)/$(APP_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "$(GREEN)✓ Installed to /usr/local/bin/$(APP_NAME)$(RESET)"
	@echo "$(YELLOW)Run '$(APP_NAME) configure' to get started$(RESET)"

uninstall: ## Uninstall system-wide
	@echo "$(BLUE)Uninstalling $(APP_NAME)...$(RESET)"
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "$(GREEN)✓ Uninstalled from /usr/local/bin/$(RESET)"

# Service management
install-service: install ## Install as system service
	@echo "$(BLUE)Installing $(APP_NAME) service...$(RESET)"
	@sudo $(APP_NAME) install
	@echo "$(GREEN)✓ Service installed$(RESET)"

uninstall-service: ## Uninstall system service
	@echo "$(BLUE)Uninstalling $(APP_NAME) service...$(RESET)"
	@sudo $(APP_NAME) uninstall || true
	@echo "$(GREEN)✓ Service uninstalled$(RESET)"

start-service: ## Start the service
	@echo "$(BLUE)Starting $(APP_NAME) service...$(RESET)"
	@sudo systemctl start $(APP_NAME) || sudo service $(APP_NAME) start
	@echo "$(GREEN)✓ Service started$(RESET)"

stop-service: ## Stop the service
	@echo "$(BLUE)Stopping $(APP_NAME) service...$(RESET)"
	@sudo systemctl stop $(APP_NAME) || sudo service $(APP_NAME) stop
	@echo "$(GREEN)✓ Service stopped$(RESET)"

status-service: ## Check service status
	@echo "$(BLUE)Checking $(APP_NAME) service status...$(RESET)"
	@sudo systemctl status $(APP_NAME) || sudo service $(APP_NAME) status

logs-service: ## View service logs
	@echo "$(BLUE)Viewing $(APP_NAME) service logs...$(RESET)"
	@sudo journalctl -u $(APP_NAME) -f || sudo tail -f /var/log/$(APP_NAME).log

# Cleanup targets
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(RESET)"

clean-all: clean ## Clean everything including vendor and cache
	@echo "$(BLUE)Cleaning everything...$(RESET)"
	@rm -rf vendor/
	@go clean -cache -modcache -testcache
	@echo "$(GREEN)✓ Deep clean complete$(RESET)"

# Development setup
setup-dev: ## Setup development environment
	@echo "$(BLUE)Setting up development environment...$(RESET)"
	@echo "$(BLUE)Installing development tools...$(RESET)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin; \
	fi
	@if ! command -v gofumpt >/dev/null 2>&1; then \
		echo "Installing gofumpt..."; \
		go install mvdan.cc/gofumpt@latest; \
	fi
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@echo "$(GREEN)✓ Development environment setup complete$(RESET)"

# Security targets
security: ## Run security checks
	@echo "$(BLUE)Running security checks...$(RESET)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(YELLOW)⚠ govulncheck not found, install with: go install golang.org/x/vuln/cmd/govulncheck@latest$(RESET)"; \
	fi
	@echo "$(GREEN)✓ Security checks complete$(RESET)"

# Release preparation
pre-release: clean deps fmt lint test security ## Prepare for release
	@echo "$(BLUE)Preparing for release...$(RESET)"
	@echo "Current version: $(VERSION)"
	@echo "Current commit: $(COMMIT)"
	@echo "$(GREEN)✓ Pre-release checks complete$(RESET)"

tag-release: ## Create and push a new release tag
	@echo "$(BLUE)Creating release tag...$(RESET)"
	@read -p "Enter version (e.g., 1.0.0): " version; \
	git tag -a "v$version" -m "Release version $version"; \
	git push origin "v$version"
	@echo "$(GREEN)✓ Release tag created and pushed$(RESET)"

# Debugging targets
debug: ## Build with debug symbols
	@echo "$(BLUE)Building debug version...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@go build -gcflags="all=-N -l" -o $(BIN_DIR)/$(APP_NAME)-debug .
	@echo "$(GREEN)✓ Debug build complete: $(BIN_DIR)/$(APP_NAME)-debug$(RESET)"

profile: ## Build with profiling enabled
	@echo "$(BLUE)Building with profiling...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@go build -tags=profile -o $(BIN_DIR)/$(APP_NAME)-profile .
	@echo "$(GREEN)✓ Profile build complete: $(BIN_DIR)/$(APP_NAME)-profile$(RESET)"

# Benchmarking
bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem ./...
	@echo "$(GREEN)✓ Benchmarks complete$(RESET)"

# Documentation
docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(RESET)"
	@go doc -all > docs/api.md
	@echo "$(GREEN)✓ Documentation generated$(RESET)"

# Information targets
info: ## Show build information
	@echo "$(BLUE)Build Information$(RESET)"
	@echo "$(BLUE)=================$(RESET)"
	@echo "App Name:     $(APP_NAME)"
	@echo "Version:      $(VERSION)"
	@echo "Commit:       $(COMMIT)"
	@echo "Build Time:   $(BUILD_TIME)"
	@echo "Go Version:   $(shell go version)"
	@echo "OS/Arch:      $(shell go env GOOS)/$(shell go env GOARCH)"
	@echo ""

env: ## Show environment information
	@echo "$(BLUE)Environment Information$(RESET)"
	@echo "$(BLUE)======================$(RESET)"
	@go env
	@echo ""

# Utility targets
size: build ## Show binary size information
	@echo "$(BLUE)Binary Size Information$(RESET)"
	@echo "$(BLUE)======================$(RESET)"
	@ls -lh $(BIN_DIR)/$(APP_NAME)
	@echo ""
	@if command -v file >/dev/null 2>&1; then \
		file $(BIN_DIR)/$(APP_NAME); \
	fi
	@if command -v upx >/dev/null 2>&1; then \
		echo "$(YELLOW)Tip: Use 'upx $(BIN_DIR)/$(APP_NAME)' to compress binary$(RESET)"; \
	fi

compress: build ## Compress binary with UPX
	@echo "$(BLUE)Compressing binary...$(RESET)"
	@if command -v upx >/dev/null 2>&1; then \
		upx --best --lzma $(BIN_DIR)/$(APP_NAME); \
		echo "$(GREEN)✓ Binary compressed$(RESET)"; \
	else \
		echo "$(RED)✗ UPX not found, install from https://upx.github.io/$(RESET)"; \
	fi

# Git hooks
install-hooks: ## Install git hooks
	@echo "$(BLUE)Installing git hooks...$(RESET)"
	@mkdir -p .git/hooks
	@echo '#!/bin/bash\nmake fmt lint test' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)✓ Git hooks installed$(RESET)"

# Quick aliases
all: release packages ## Build everything
quick: fmt test build ## Quick development build
ci: deps lint test build ## CI pipeline simulation

# Help for specific categories
help-build: ## Show build-related targets
	@echo "$(BLUE)Build Targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^build.*:.*?## / {printf "  $(YELLOW)%-15s$(RESET) %s\n", $1, $2}' $(MAKEFILE_LIST)

help-dev: ## Show development-related targets
	@echo "$(BLUE)Development Targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^(dev|test|lint|fmt|vet|debug|bench).*:.*?## / {printf "  $(YELLOW)%-15s$(RESET) %s\n", $1, $2}' $(MAKEFILE_LIST)

help-deploy: ## Show deployment-related targets
	@echo "$(BLUE)Deployment Targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^(install|uninstall|release|packages|docker).*:.*?## / {printf "  $(YELLOW)%-15s$(RESET) %s\n", $1, $2}' $(MAKEFILE_LIST)

# Version information for the binary
version:
	@echo $(VERSION)

# Check if we're on macOS and provide specific instructions
check-macos: ## Check macOS compatibility and show setup instructions
	@if [ "$(uname)" = "Darwin" ]; then \
		echo "$(GREEN)✓ macOS detected$(RESET)"; \
		echo "$(BLUE)macOS Setup Instructions:$(RESET)"; \
		echo "1. Build: make build"; \
		echo "2. Test: ./bin/serverhealth configure"; \
		echo "3. Run: ./bin/serverhealth start"; \
		echo "4. Install: make install-local"; \
		echo "5. Service: serverhealth install"; \
		echo ""; \
		echo "$(YELLOW)Note: Service will use launchd on macOS$(RESET)"; \
	else \
		echo "$(YELLOW)⚠ Not running on macOS$(RESET)"; \
	fi

# macOS specific targets
macos-build: ## Build specifically for macOS
	@echo "$(BLUE)Building for macOS (current architecture)...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@go build $(BUILD_FLAGS) -o $(BIN_DIR)/serverhealth .
	@echo "$(GREEN)✓ macOS build complete: $(BIN_DIR)/serverhealth$(RESET)"

macos-test: macos-build ## Build and test on macOS
	@echo "$(BLUE)Testing serverhealth on macOS...$(RESET)"
	@./$(BIN_DIR)/serverhealth --version
	@echo "$(GREEN)✓ Basic test passed$(RESET)"
	@echo "$(YELLOW)Run './$(BIN_DIR)/serverhealth configure' to set up$(RESET)"

macos-install: macos-build ## Install for current macOS user
	@echo "$(BLUE)Installing serverhealth for current user...$(RESET)"
	@mkdir -p ~/bin
	@cp $(BIN_DIR)/serverhealth ~/bin/
	@chmod +x ~/bin/serverhealth
	@echo "$(GREEN)✓ Installed to ~/bin/serverhealth$(RESET)"
	@echo "$(YELLOW)Add ~/bin to your PATH if not already done$(RESET)"
	@echo "$(YELLOW)Run 'serverhealth configure' to get started$(RESET)"

# Quick start for new users
quickstart: ## Quick start guide
	@echo "$(BLUE)ServerHealth Quick Start Guide$(RESET)"
	@echo "$(BLUE)=============================$(RESET)"
	@echo ""
	@echo "$(GREEN)1. Build the application:$(RESET)"
	@echo "   make build"
	@echo ""
	@echo "$(GREEN)2. Configure monitoring:$(RESET)"
	@echo "   ./bin/serverhealth configure"
	@echo ""
	@echo "$(GREEN)3. Start monitoring:$(RESET)"
	@echo "   ./bin/serverhealth start"
	@echo ""
	@echo "$(GREEN)4. Install as service (optional):$(RESET)"
	@echo "   make install"
	@echo "   sudo serverhealth install"
	@echo ""
	@echo "$(BLUE)For macOS users:$(RESET)"
	@echo "   make macos-build"
	@echo "   make macos-install"
	@echo ""

# Development workflow
dev-start: ## Complete development start workflow
	@echo "$(BLUE)Starting development workflow...$(RESET)"
	@make deps
	@make fmt
	@make build
	@echo "$(GREEN)✓ Ready for development!$(RESET)"
	@echo "$(YELLOW)Next: ./bin/serverhealth configure$(RESET)"

# Verification targets
verify-build: build ## Verify the build works
	@echo "$(BLUE)Verifying build...$(RESET)"
	@if [ -f "$(BIN_DIR)/serverhealth" ]; then \
		echo "$(GREEN)✓ Binary exists$(RESET)"; \
		$(BIN_DIR)/serverhealth --version || echo "$(YELLOW)Version check failed$(RESET)"; \
		$(BIN_DIR)/serverhealth --help | head -5; \
		echo "$(GREEN)✓ Build verification complete$(RESET)"; \
	else \
		echo "$(RED)✗ Binary not found$(RESET)"; \
		exit 1; \
	fi

# File watching for development (requires fswatch on macOS)
watch: ## Watch files and rebuild on changes (macOS)
	@if command -v fswatch >/dev/null 2>&1; then \
		echo "$(BLUE)Watching for file changes...$(RESET)"; \
		fswatch -o . | xargs -n1 -I{} make build; \
	else \
		echo "$(YELLOW)⚠ fswatch not found. Install with: brew install fswatch$(RESET)"; \
	fi

# Show dependency versions
deps-info: ## Show dependency information
	@echo "$(BLUE)Dependency Information$(RESET)"
	@echo "$(BLUE)=====================$(RESET)"
	@go list -m -versions all | head -10
	@echo ""
	@echo "$(BLUE)Direct dependencies:$(RESET)"
	@go mod graph | grep "^$(shell go list -m)" | cut -d' ' -f2 | sort | uniq

# Final target to show completion
complete: ## Mark task as complete
	@echo "$(GREEN)"
	@echo "  ✓ ServerHealth build system ready!"
	@echo "  ✓ Run 'make quickstart' for usage guide"
	@echo "  ✓ Run 'make help' for all available commands"
	@echo "$(RESET)"
