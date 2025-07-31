# PlaylistSync Makefile
.PHONY: build build-em run run-dev dev clean lint fix test deps mocks help
.PHONY: frontend-install frontend-dev frontend-build
.PHONY: build-all run-prod

# Default target
help:
	@echo "Available commands:"
	@echo ""
	@echo "Backend:"
	@echo "  build      - Build the Go application"
	@echo "  run        - Run the application in production mode"
	@echo "  run-dev    - Run the application in development mode"
	@echo "  dev        - Run the application in development mode with hot reload (air)"
	@echo "  clean      - Clean build artifacts"
	@echo "  lint       - Run golangci-lint to check code quality"
	@echo "  fix        - Format and fix code issues"
	@echo "  test       - Run tests"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  mocks      - Generate all mocks using go generate"
	@echo ""
	@echo "Frontend:"
	@echo "  frontend-install  - Install frontend dependencies"
	@echo "  frontend-dev      - Start frontend development server"
	@echo "  frontend-build    - Build frontend for production"
	@echo ""
	@echo "Full Stack:"
	@echo "  build-all     - Build both frontend and backend for production"
	@echo "  run-prod 	   - Build everything and run in production mode" 
	@echo "  dev-full      - Start both backend and frontend in development mode"
	@echo "  help          - Show this help message"

# Build the application
build:
	@echo "Building application..."
	go build -o playlist-router ./cmd/pb

# Build the application with the frontend embedded
build-all: frontend-build
	@echo "Building application with embedded frontend..."
	@mkdir -p internal/static/dist
	@cp -r web/dist/* internal/static/dist/
	go build -o playlist-router ./cmd/pb
	@echo "Full stack build completed!"

# Run in production mode
run: build
	@echo "Starting server in production mode..."
	./playlist-router serve

# Run in dev mode
run-dev: build
	@echo "Starting server in development mode..."
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

# Frontend commands
frontend-install:
	@echo "Installing frontend dependencies..."
	cd web && npm install

frontend-dev:
	@echo "Starting frontend development server..."
	cd web && npm run dev

frontend-build:
	@echo "Building frontend for production..."
	cd web && npm run build

# Build everything and run in production mode
run-prod: build-all
	@echo "Starting production server with integrated frontend..."
	@echo "Server will be available at http://localhost:8090"
	./playlist-router serve

# Full stack development
dev-full:
	@echo "Starting full stack development..."
	@echo "Backend will be available at http://localhost:8090"
	@echo "Frontend will be available at http://localhost:5173"
	@echo "Press Ctrl+C to stop both servers"
	@make -j2 dev frontend-dev