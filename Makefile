# Makefile for Go version of Humanity simulation

# Build targets
.PHONY: all build clean run test

# Default target
all: build

# Build the application
build:
	go build -o humanity .

# Build with optimizations
build-release:
	go build -ldflags="-s -w" -o humanity .

# Run the application
run: build
	./humanity

# Clean build artifacts
clean:
	rm -f humanity
	rm -f logs_*.txt

# Run tests (if any)
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod tidy
	go mod download

# Cross-compile for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o humanity-linux .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o humanity.exe .

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o humanity-mac .

# Build for all platforms
build-all: build-linux build-windows build-mac

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-release - Build with optimizations"
	@echo "  run           - Build and run the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Install dependencies"
	@echo "  build-all     - Cross-compile for all platforms"
	@echo "  help          - Show this help"