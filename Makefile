.PHONY: build test clean install deps fmt vet lint

# Build variables
BINARY_NAME_DAEMON=jtnt-agentd
BINARY_NAME_CLI=jtnt-agent
BUILD_DIR=bin
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# Default target
all: build

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Build both daemon and CLI
build: build-daemon build-cli

# Build daemon
build-daemon:
	@echo "Building daemon..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_DAEMON) ./cmd/agentd

# Build CLI
build-cli:
	@echo "Building CLI..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_CLI) ./cmd/jtnt-agent

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME_DAEMON) ./cmd/agentd
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/linux/$(BINARY_NAME_CLI) ./cmd/jtnt-agent

# Build for macOS
build-darwin:
	@echo "Building for macOS (Intel)..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/darwin-amd64/$(BINARY_NAME_DAEMON) ./cmd/agentd
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/darwin-amd64/$(BINARY_NAME_CLI) ./cmd/jtnt-agent
	@echo "Building for macOS (Apple Silicon)..."
	@mkdir -p $(BUILD_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/darwin-arm64/$(BINARY_NAME_DAEMON) ./cmd/agentd
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/darwin-arm64/$(BINARY_NAME_CLI) ./cmd/jtnt-agent

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/windows/$(BINARY_NAME_DAEMON).exe ./cmd/agentd
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/windows/$(BINARY_NAME_CLI).exe ./cmd/jtnt-agent

# Run tests
test:
	@echo "Running tests..."
	$(GO) test ./... -v -race -coverprofile=coverage.out

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install from https://golangci-lint.run/"; exit 1; }
	golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install binaries to system (requires sudo on Linux/macOS)
install: build
	@echo "Installing binaries..."
ifeq ($(shell uname -s),Linux)
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME_DAEMON) /usr/local/bin/
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME_CLI) /usr/local/bin/
	@echo "Installed to /usr/local/bin/"
else ifeq ($(shell uname -s),Darwin)
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME_DAEMON) /usr/local/bin/
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME_CLI) /usr/local/bin/
	@echo "Installed to /usr/local/bin/"
else
	@echo "Manual installation required on Windows"
	@echo "Copy $(BUILD_DIR)/*.exe to desired location"
endif

# Uninstall binaries from system
uninstall:
	@echo "Uninstalling binaries..."
ifeq ($(shell uname -s),Linux)
	sudo rm -f /usr/local/bin/$(BINARY_NAME_DAEMON)
	sudo rm -f /usr/local/bin/$(BINARY_NAME_CLI)
else ifeq ($(shell uname -s),Darwin)
	sudo rm -f /usr/local/bin/$(BINARY_NAME_DAEMON)
	sudo rm -f /usr/local/bin/$(BINARY_NAME_CLI)
else
	@echo "Manual uninstallation required on Windows"
endif

# Development run (build and run daemon)
run: build-daemon
	@echo "Running daemon..."
	sudo $(BUILD_DIR)/$(BINARY_NAME_DAEMON)

# Help target
help:
	@echo "JTNT Agent Build System"
	@echo ""
	@echo "Targets:"
	@echo "  build          - Build both daemon and CLI for current platform"
	@echo "  build-all      - Build for all platforms (Linux, macOS, Windows)"
	@echo "  build-linux    - Build for Linux"
	@echo "  build-darwin   - Build for macOS (Intel and Apple Silicon)"
	@echo "  build-windows  - Build for Windows"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests and generate coverage report"
	@echo "  fmt            - Format Go code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run golangci-lint"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binaries to system"
	@echo "  uninstall      - Remove binaries from system"
	@echo "  run            - Build and run daemon"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  help           - Show this help message"
