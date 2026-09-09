# Aranet4 Go Reader - Makefile

# Project info
BINARY_NAME=aranet4-go
VERSION?=1.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DEFAULT_MAC=$(shell test -f DEFAULT-MAC-ADDR && cat DEFAULT-MAC-ADDR | tr -d '\n' || echo "")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.DefaultMAC=$(DEFAULT_MAC)"
BUILDFLAGS=#-a
CGO_ENABLED=0

# Output directories
BUILD_DIR=build
DIST_DIR=dist

.PHONY: all build clean test deps install uninstall run help \
        build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows \
        build-all release

# Default target
all: deps build

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## deps: Download and install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v

## build-linux: Build for Linux (amd64)
build-linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v

## build-linux-arm64: Build for Linux ARM64 (Raspberry Pi)
build-linux-arm64:
	@echo "Building for Linux (arm64)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -v

## build-darwin: Build for macOS (Intel)
build-darwin:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v

## build-darwin-arm64: Build for macOS (Apple Silicon)
build-darwin-arm64:
	@echo "Building for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -v

## build-windows: Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v

## build-all: Build for all supported platforms (Linux + macOS; Windows not supported by go-ble/ble)
build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64
	@echo "All builds complete!"
	@ls -lh $(BUILD_DIR)

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)

## install: Install binary to system (requires sudo on Linux)
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Setting Bluetooth capabilities..."
	@sudo setcap 'cap_net_raw,cap_net_admin=eip' /usr/local/bin/$(BINARY_NAME) || true
	@echo "Installing man page..."
	@sudo mkdir -p /usr/local/share/man/man1
	@sudo cp $(BINARY_NAME).1 /usr/local/share/man/man1/
	@sudo gzip -f /usr/local/share/man/man1/$(BINARY_NAME).1
	@echo "Installation complete!"

## uninstall: Remove binary from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@sudo rm -f /usr/local/share/man/man1/$(BINARY_NAME).1.gz
	@echo "Uninstall complete!"

## run: Build and run with example MAC (change MAC in Makefile or use: make run MAC=XX:XX:XX:XX:XX:XX)
MAC?=FC:5C:65:B7:84:94
run: build
	@echo "Running $(BINARY_NAME) with MAC=$(MAC)..."
	@$(BUILD_DIR)/$(BINARY_NAME) -mac $(MAC)

## run-json: Build and run with JSON output
run-json: build
	@echo "Running $(BINARY_NAME) with JSON output..."
	@$(BUILD_DIR)/$(BINARY_NAME) -mac $(MAC) -json

## release: Create release packages for all platforms
release: clean build-all
	@echo "Creating release packages..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && \
		tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip -q ../$(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@cd $(DIST_DIR) && sha256sum * > checksums.txt
	@echo "Release packages created in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

## mod-update: Update Go module dependencies
mod-update:
	@echo "Updating dependencies..."
	$(GOMOD) get -u
	$(GOMOD) tidy

## info: Display build information
info:
	@echo "Binary Name:  $(BINARY_NAME)"
	@echo "Version:      $(VERSION)"
	@echo "Build Time:   $(BUILD_TIME)"
	@echo "Git Commit:   $(GIT_COMMIT)"
	@echo "Go Version:   $(shell $(GOCMD) version)"
