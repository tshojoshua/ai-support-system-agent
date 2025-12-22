.PHONY: build test clean install build-all

BINARY_NAME=jtnt-agentd
CLI_NAME=jtnt-agent
VERSION=1.0.0

# Build for current platform
build:
	@echo "Building for current platform..."
	go build -o bin/$(BINARY_NAME) -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/agentd
	go build -o bin/$(CLI_NAME) -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/jtnt-agent

# Build for all platforms
build-all:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/agentd
	GOOS=windows GOARCH=amd64 go build -o bin/$(CLI_NAME)-windows-amd64.exe -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/jtnt-agent
	
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/agentd
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/agentd
	GOOS=darwin GOARCH=amd64 go build -o bin/$(CLI_NAME)-darwin-amd64 -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/jtnt-agent
	GOOS=darwin GOARCH=arm64 go build -o bin/$(CLI_NAME)-darwin-arm64 -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/jtnt-agent
	
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/agentd
	GOOS=linux GOARCH=amd64 go build -o bin/$(CLI_NAME)-linux-amd64 -ldflags "-s -w -X github.com/jtnt/agent/internal/config.Version=$(VERSION)" ./cmd/jtnt-agent

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Install (Linux/macOS)
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	sudo cp bin/$(CLI_NAME) /usr/local/bin/
	@echo "Installation complete!"

# Development run
dev: build
	./bin/$(BINARY_NAME)
