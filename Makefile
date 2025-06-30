COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html
MIN_COVERAGE  := 70

# Go source files (excluding vendor)
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v -coverprofile=$(COVERAGE_FILE) ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@golangci-lint run --timeout 5m

.PHONY: coverage
coverage:
	@echo "Generating test coverage report..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage HTML report generated: $(COVERAGE_HTML)"
	@open $(COVERAGE_HTML)

.PHONY: coverage-check
coverage-check:
	@COVERAGE=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	RESULT=$$(echo "$$COVERAGE < $(MIN_COVERAGE)" | bc); \
	if [ "$$RESULT" -eq 1 ]; then \
		echo "Coverage is below $(MIN_COVERAGE)%: $$COVERAGE%"; \
		exit 1; \
	else \
		echo "Coverage is sufficient: $$COVERAGE%"; \
	fi