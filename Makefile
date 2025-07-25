# PlaylistSync Makefile
.PHONY: build run run-dev dev clean lint fix test deps mocks help

# Default target
help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application in production mode"
	@echo "  run-dev  - Run the application in development mode"
	@echo "  dev      - Run the application in development mode with hot reload (air)"
	@echo "  clean    - Clean build artifacts"
	@echo "  lint     - Run golangci-lint to check code quality"
	@echo "  fix      - Format and fix code issues"
	@echo "  test     - Run tests"
	@echo "  deps     - Download and tidy dependencies"
	@echo "  mocks    - Generate all mocks using go generate"
	@echo "  help     - Show this help message"

# Build the application
build:
	@echo "Building application..."
	go build -o playlist-router ./cmd/pb

# Run in production mode
run: build
	@echo "Starting server in production mode..."
	./playlist-router serve

# Run in production mode
run-dev: build
	@echo "Starting server in production mode..."
	./playlist-router serve --dev

# Run in development mode with hot reload
dev:
	@echo "Starting development server with hot reload..."
	air

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f pb/main
	rm -rf tmp/

# Run linter
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Format and fix code issues
fix:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .
	@echo "Running golangci-lint --fix..."
	golangci-lint run --fix

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Generate all mocks
mocks:
	@echo "Generating mocks..."
	go generate ./...