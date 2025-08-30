COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

FORMAT ?= dots

.PHONY: fmt
fmt: # Format code
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: lint
lint: # Lint code
	@echo "Linting code..."
	@golangci-lint run --timeout 5m

.PHONY: install
install: # Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

install-tools: # Install tools
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@go install gotest.tools/gotestsum@latest

.PHONY: test
test: # Run tests
	@echo "Running tests..."
	@gotestsum --format=$(FORMAT) -- ./...

.PHONY: testcov
testcov: # Run tests with coverage
	@echo "Running tests with coverage..."
	@gotestsum --format=$(FORMAT) -- -coverprofile=$(COVERAGE_FILE) ./...
	@echo "Total coverage is: " $(shell go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}')

.PHONY: testcov-html
testcov-html: testcov # Generate coverage HTML report
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage HTML report generated: $(COVERAGE_HTML)"
	@open $(COVERAGE_HTML)

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  install      - Install dependencies"
	@echo "  install-tools- Install development tools"
	@echo "  test         - Run tests"
	@echo "  testcov      - Run tests with coverage"
	@echo "  testcov-html - Generate coverage HTML report"
	@echo "  help         - Show this help message"