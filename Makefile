COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Go source files (excluding vendor)
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@go vet ./...

.PHONY: coverage
coverage:
	@echo "Generating test coverage report..."
	@go test -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage HTML report generated: $(COVERAGE_HTML)"
	@open $(COVERAGE_HTML)