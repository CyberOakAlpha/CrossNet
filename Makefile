.PHONY: build clean test install windows linux darwin all

# Build variables
BINARY_NAME=crossnet
GUI_BINARY_NAME=crossnet-gui
BUILD_DIR=build
VERSION=1.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: clean build build-gui

# Build CLI for current platform
build:
	@echo "Building CrossNet CLI for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/crossnet/main.go

# Build GUI for current platform
build-gui:
	@echo "Building CrossNet GUI for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME) cmd/crossnet-gui/main.go

# Build for Windows (AMD64 and ARM64)
windows:
	@echo "Building CrossNet for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/crossnet/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME)-windows-amd64.exe cmd/crossnet-gui/main.go
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe cmd/crossnet/main.go
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME)-windows-arm64.exe cmd/crossnet-gui/main.go

# Build for Linux (AMD64 and ARM64)
linux:
	@echo "Building CrossNet for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/crossnet/main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME)-linux-amd64 cmd/crossnet-gui/main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 cmd/crossnet/main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME)-linux-arm64 cmd/crossnet-gui/main.go
	GOOS=linux GOARCH=arm go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv7 cmd/crossnet/main.go
	GOOS=linux GOARCH=arm go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME)-linux-armv7 cmd/crossnet-gui/main.go

# Build for macOS (AMD64 and ARM64 - Apple Silicon)
darwin:
	@echo "Building CrossNet for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/crossnet/main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME)-darwin-amd64 cmd/crossnet-gui/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/crossnet/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY_NAME)-darwin-arm64 cmd/crossnet-gui/main.go

# Build for all platforms
cross-compile: windows linux darwin
	@echo "Cross-compilation completed for all platforms"

# Generate checksums
checksums:
	@echo "Generating SHA256 checksums..."
	@cd $(BUILD_DIR) && sha256sum * > SHA256SUMS
	@echo "Checksums saved to $(BUILD_DIR)/SHA256SUMS"

# Build everything and generate checksums
release: clean cross-compile checksums
	@echo "Release build completed with checksums"

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
	@echo "  build-gui     - Build GUI for current platform"
	@echo "  windows       - Build for Windows (amd64, arm64)"
	@echo "  linux         - Build for Linux (amd64, arm64, armv7)"
	@echo "  darwin        - Build for macOS (amd64, arm64)"
	@echo "  cross-compile - Build for all platforms"
	@echo "  checksums     - Generate SHA256 checksums"
	@echo "  release       - Full release build with checksums"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build directory"
	@echo "  install       - Install to system (requires sudo)"
	@echo "  run           - Run without building"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  help          - Show this help"