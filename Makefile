.PHONY: test test-unit test-coverage test-race test-verbose clean-coverage help

# Default target
test: test-unit

# Run all unit tests
test-unit:
	@echo "Running unit tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage summary:"
	go tool cover -func=coverage.out | grep total

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test ./... -race

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test ./... -v -cover

# Run tests for a specific package
test-package:
	@if [ -z "$(PKG)" ]; then echo "Usage: make test-package PKG=internal/app"; exit 1; fi
	go test ./$(PKG) -v -cover

# Clean coverage files
clean-coverage:
	@echo "Cleaning coverage files..."
	rm -f coverage.out coverage.html

# Run tests and open coverage report
test-coverage-open: test-coverage
	@if command -v open >/dev/null 2>&1; then \
		open coverage.html; \
	elif command -v xdg-open >/dev/null 2>&1; then \
		xdg-open coverage.html; \
	else \
		echo "Coverage report saved to coverage.html"; \
	fi

# Build the application
build:
	@echo "Building application..."
	go build -o docker-util .

# Run the application
run:
	@echo "Running application..."
	go run .

# Show help
help:
	@echo "Available targets:"
	@echo "  test              - Run unit tests (default)"
	@echo "  test-unit         - Run unit tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-race         - Run tests with race detection"
	@echo "  test-verbose      - Run tests with verbose output"
	@echo "  test-package      - Run tests for specific package (use PKG=path)"
	@echo "  test-coverage-open - Run coverage and open report"
	@echo "  clean-coverage    - Clean coverage files"
	@echo "  build             - Build the application"
	@echo "  run               - Run the application"
	@echo "  help              - Show this help" 