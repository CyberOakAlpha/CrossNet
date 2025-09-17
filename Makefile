.PHONY: build clean test install windows linux darwin all

# Build variables
BINARY_NAME=crossnet
BUILD_DIR=build
VERSION=1.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: clean build

# Build for current platform
build:
	@echo "Building CrossNet for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/crossnet/main.go

# Build for Windows
windows:
	@echo "Building CrossNet for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/crossnet/main.go

# Build for Linux
linux:
	@echo "Building CrossNet for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/crossnet/main.go

# Build for macOS
darwin:
	@echo "Building CrossNet for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/crossnet/main.go

# Build for all platforms
cross-compile: windows linux darwin
	@echo "Cross-compilation completed for all platforms"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build directory
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Install to system (Unix-like systems)
install: build
	@echo "Installing CrossNet to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "CrossNet installed successfully!"

# Development run
run:
	@go run cmd/crossnet/main.go

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build for current platform"
	@echo "  windows       - Build for Windows"
	@echo "  linux         - Build for Linux"
	@echo "  darwin        - Build for macOS"
	@echo "  cross-compile - Build for all platforms"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build directory"
	@echo "  install       - Install to system (requires sudo)"
	@echo "  run           - Run without building"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  help          - Show this help"