.PHONY: build test lint clean run fmt deps build-all build-mac-universal changelog help

# Variables
BINARY_NAME=zeno
BUILD_DIR=bin
# Version from git tag (strip 'v' prefix). Override with TAG= or VERSION=.
TAG ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "")
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null | sed 's/^v//' || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X github.com/zenlayer/zenlayercloud-cli/internal/version.Version=$(VERSION) \
  -X github.com/zenlayer/zenlayercloud-cli/internal/version.GitCommit=$(GIT_COMMIT) \
  -X github.com/zenlayer/zenlayercloud-cli/internal/version.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build the application
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Run linters
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	@if command -v goimports > /dev/null; then goimports -w .; fi

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Mac Universal Binary (fat binary for Intel + Apple Silicon, requires macOS with lipo)
build-mac-universal:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	lipo -create -output $(BUILD_DIR)/$(BINARY_NAME)-darwin-universal \
		$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 \
		$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	@echo "Universal binary: $(BUILD_DIR)/$(BINARY_NAME)-darwin-universal"

# Install locally
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Run with race detector
run-race:
	go run -race .

# Generate CHANGELOG.md from git history (requires git-cliff)
changelog:
	git-cliff -o CHANGELOG.md

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests with race detector"
	@echo "  test-coverage - Run tests and generate coverage report"
	@echo "  lint          - Run linters (requires golangci-lint)"
	@echo "  fmt           - Format code"
	@echo "  run           - Build and run the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  build-all          - Build for all platforms"
	@echo "  build-mac-universal - Build Mac Universal binary (Intel + Apple Silicon)"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  changelog     - Generate CHANGELOG.md (requires git-cliff)"
	@echo "  help          - Show this help"
