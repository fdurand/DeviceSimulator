# DeviceSimulator Build Configuration

# Build with optimizations
.PHONY: build clean test lint run debug profile

# Variables
BINARY_NAME=device-simulator
BINARY_PATH=./bin/$(BINARY_NAME)
CONFIG_FILE=config.ini
PKG=./...

# Default target
all: clean test build

# Build optimized binary
build:
	@echo "Building optimized binary..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build \
		-ldflags='-w -s -extldflags "-static"' \
		-a -installsuffix cgo \
		-o $(BINARY_PATH) .
	@echo "Binary built: $(BINARY_PATH)"

# Build with debug symbols
debug:
	@echo "Building debug binary..."
	@mkdir -p bin
	go build -race -o $(BINARY_PATH) .
	@echo "Debug binary built: $(BINARY_PATH)"

# Build with profiling support
profile:
	@echo "Building profile binary..."
	@mkdir -p bin
	go build -o $(BINARY_PATH) -tags profile .
	@echo "Profile binary built: $(BINARY_PATH)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "Clean completed"

# Run tests
test:
	@echo "Running tests..."
	@go test -v $(PKG)
	@echo "Tests completed"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	@go test -race -v $(PKG)
	@echo "Race tests completed"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. $(PKG)
	@echo "Benchmarks completed"

# Lint code
lint:
	@echo "Running linter..."
	@go vet $(PKG)
	@go fmt $(PKG)
	@echo "Linting completed"

# Run the simulator
run: build
	@echo "Running DeviceSimulator..."
	@sudo $(BINARY_PATH) -file $(CONFIG_FILE)

# Run Xerox printer simulation
run-xerox: build
	@echo "Running Xerox Printer Simulation..."
	@sudo $(BINARY_PATH) -file config-xerox-printer.ini

# Run Xerox simulation with debug
debug-xerox: debug
	@echo "Running Xerox Printer Simulation (Debug Mode)..."
	@sudo $(BINARY_PATH) -file config-xerox-printer.ini -debug

# Install Go tools
tools:
	@echo "Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "Tools installed"

# Update dependencies
deps:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies updated"

# Generate code coverage report
coverage:
	@echo "Generating coverage report..."
	@mkdir -p coverage
	@go test -coverprofile=coverage/coverage.out $(PKG)
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"

# Docker build (if needed)
docker:
	@echo "Building Docker image..."
	@docker build -t device-simulator .
	@echo "Docker image built: device-simulator"

# Validate configurations
validate:
	@echo "Validating configurations..."
	@chmod +x validate-xerox-config.sh
	@./validate-xerox-config.sh
	@echo "Validation completed"

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build optimized binary"
	@echo "  debug       - Build debug binary with race detection"
	@echo "  profile     - Build binary with profiling support"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  test-race   - Run tests with race detection"
	@echo "  bench       - Run benchmarks"
	@echo "  lint        - Lint code"
	@echo "  run         - Build and run default simulator"
	@echo "  run-xerox   - Run Xerox printer simulation"
	@echo "  debug-xerox - Run Xerox simulation with debug"
	@echo "  validate    - Validate configuration files"
	@echo "  tools       - Install development tools"
	@echo "  deps        - Update dependencies"
	@echo "  coverage    - Generate test coverage report"
	@echo "  docker      - Build Docker image"
	@echo "  help        - Show this help"