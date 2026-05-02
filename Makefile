COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

FORMAT ?= dots
VERSION ?=
PUSH ?= false
VERSION_STRIPPED := $(patsubst v%,%,$(VERSION))
VERSION_TAG := v$(VERSION_STRIPPED)
BLUE := \033[0;34m
YELLOW := \033[1;33m
NC := \033[0m

ROOT_MODULE := .
EXTRA_MODULES := extra/migriscli extra/migriscobra
MONOREPO_MODULES := $(ROOT_MODULE) $(EXTRA_MODULES)
RELEASE_TAGS := $(VERSION_TAG) extra/migriscli/$(VERSION_TAG) extra/migriscobra/$(VERSION_TAG)

define run_in_modules
	@set -e; \
	for mod in $(MONOREPO_MODULES); do \
		echo "==> [$${mod}] $(1)"; \
		(cd "$$mod" && $(2)); \
	done
endef

.PHONY: fmt
fmt: # Format code
	$(call run_in_modules,Formatting code...,go fmt ./...)

.PHONY: lint
lint: # Lint code
	$(call run_in_modules,Linting code...,golangci-lint run --timeout=2m --verbose)

.PHONY: lint-fix
lint-fix: # Fix lint issues
	$(call run_in_modules,Fixing lint issues...,golangci-lint run --fix --timeout=2m --verbose)

.PHONY: install
install: # Install dependencies
	$(call run_in_modules,Installing dependencies...,go mod download && go mod tidy)

.PHONY: tidy
tidy: # Run go mod tidy in all modules
	$(call run_in_modules,Tidying modules...,go mod tidy)

install-tools: # Install tools
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@go install gotest.tools/gotestsum@latest

.PHONY: test
test: # Run tests
	$(call run_in_modules,Running tests...,gotestsum --format=$(FORMAT) -- ./...)

.PHONY: testcov
testcov: # Run tests with coverage
	@set -e; \
	for mod in $(MONOREPO_MODULES); do \
		echo "==> [$${mod}] Running tests with coverage..."; \
		(cd "$$mod" && gotestsum --format=$(FORMAT) -- -coverprofile=$(COVERAGE_FILE) ./...); \
		echo "==> [$${mod}] Total coverage is: $$(cd "$$mod" && go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}')"; \
	done

.PHONY: testcov-html
testcov-html: testcov # Generate coverage HTML report
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage HTML report generated: $(COVERAGE_HTML)"
	@open $(COVERAGE_HTML)

.PHONY: sync-extra-deps
sync-extra-deps: # Sync extra modules to require github.com/akfaiz/migris@$(VERSION_TAG)
	@set -e; \
	if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required. Usage: make sync-extra-deps VERSION=0.5.0 (or v0.5.0)"; \
		exit 1; \
	fi; \
	if ! printf '%s' "$(VERSION_STRIPPED)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$'; then \
		echo "VERSION must be SemVer (example: 0.5.0 or v0.5.0)"; \
		exit 1; \
	fi; \
	for mod in $(EXTRA_MODULES); do \
		echo "Syncing $$mod to github.com/akfaiz/migris@$(VERSION_TAG)"; \
		(cd "$$mod" && go mod edit -require=github.com/akfaiz/migris@$(VERSION_TAG) && go mod tidy); \
	done

.PHONY: release-core-preflight
release-core-preflight: # Validate core tag release prerequisites
	@set -e; \
	if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required. Usage: make release-core VERSION=0.5.0 (or v0.5.0)"; \
		exit 1; \
	fi; \
	if ! printf '%s' "$(VERSION_STRIPPED)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$'; then \
		echo "VERSION must be SemVer (example: 0.5.0 or v0.5.0)"; \
		exit 1; \
	fi; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working tree must be clean before release."; \
		exit 1; \
	fi; \
	if ! git rev-parse --abbrev-ref --symbolic-full-name @{u} >/dev/null 2>&1; then \
		echo "No upstream branch configured."; \
		exit 1; \
	fi; \
	if [ "$$(git rev-list --count @{u}..HEAD)" -ne 0 ]; then \
		echo "You have unpushed commits. Push before release."; \
		exit 1; \
	fi; \
	if git rev-parse -q --verify "refs/tags/$(VERSION_TAG)" >/dev/null; then \
		echo "Tag already exists locally: $(VERSION_TAG)"; \
		exit 1; \
	fi; \
	if git ls-remote --exit-code --tags origin "refs/tags/$(VERSION_TAG)" >/dev/null 2>&1; then \
		echo "Tag already exists on origin: $(VERSION_TAG)"; \
		exit 1; \
	fi

.PHONY: release-core-publish
release-core-publish: release-core-preflight # Create and push root module tag
	@set -e; \
	echo "Creating core tag $(VERSION_TAG)"; \
	git tag -a "$(VERSION_TAG)" -m "release $(VERSION_TAG)"; \
	echo "Pushing core tag $(VERSION_TAG)"; \
	git push origin "$(VERSION_TAG)"

.PHONY: release-extra-preflight
release-extra-preflight: # Validate extra module release prerequisites
	@set -e; \
	if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required. Usage: make release-extra VERSION=0.5.0 (or v0.5.0)"; \
		exit 1; \
	fi; \
	if ! printf '%s' "$(VERSION_STRIPPED)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$'; then \
		echo "VERSION must be SemVer (example: 0.5.0 or v0.5.0)"; \
		exit 1; \
	fi; \
	if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working tree must be clean before publishing extra module tags."; \
		exit 1; \
	fi; \
	if ! git rev-parse --abbrev-ref --symbolic-full-name @{u} >/dev/null 2>&1; then \
		echo "No upstream branch configured."; \
		exit 1; \
	fi; \
	if [ "$$(git rev-list --count @{u}..HEAD)" -ne 0 ]; then \
		echo "You have unpushed commits. Push before publishing extra tags."; \
		exit 1; \
	fi; \
	for mod in $(EXTRA_MODULES); do \
		current="$$(cd "$$mod" && go list -m -f '{{.Version}}' github.com/akfaiz/migris)"; \
		if [ "$$current" != "$(VERSION_TAG)" ]; then \
			echo "$$mod requires github.com/akfaiz/migris@$$current (expected $(VERSION_TAG))."; \
			echo "Run: make sync-extra-deps VERSION=$(VERSION_TAG)"; \
			exit 1; \
		fi; \
	done; \
	for tag in extra/migriscli/$(VERSION_TAG) extra/migriscobra/$(VERSION_TAG); do \
		if git rev-parse -q --verify "refs/tags/$$tag" >/dev/null; then \
			echo "Tag already exists locally: $$tag"; \
			exit 1; \
		fi; \
		if git ls-remote --exit-code --tags origin "refs/tags/$$tag" >/dev/null 2>&1; then \
			echo "Tag already exists on origin: $$tag"; \
			exit 1; \
		fi; \
	done

.PHONY: release-extra-publish
release-extra-publish: release-extra-preflight # Create and push extra module tags
	@set -e; \
	tags="extra/migriscli/$(VERSION_TAG) extra/migriscobra/$(VERSION_TAG)"; \
	for tag in $$tags; do \
		echo "Creating extra tag $$tag"; \
		git tag -a "$$tag" -m "release $$tag"; \
	done; \
	echo "Pushing extra tags"; \
	git push origin $$tags

.PHONY: release-core
release-core: # Stage 1: publish root tag and sync extra deps
	@set -e; \
	$(MAKE) release-core-publish VERSION=$(VERSION_TAG); \
	$(MAKE) sync-extra-deps VERSION=$(VERSION_TAG); \
	echo "Stage 1 complete."; \
	echo "Commit and push extra dependency updates, then run:"; \
	echo "  make release-extra VERSION=$(VERSION_TAG)"

.PHONY: release-extra
release-extra: # Stage 2: publish extra module tags after sync commit is pushed
	@set -e; \
	$(MAKE) release-extra-publish VERSION=$(VERSION_TAG)

.PHONY: release-dry-run
release-dry-run: # Dry-run for staged monorepo release
	@set -e; \
	if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required. Usage: make release-dry-run VERSION=0.5.0 (or v0.5.0)"; \
		exit 1; \
	fi; \
	if ! printf '%s' "$(VERSION_STRIPPED)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$'; then \
		echo "VERSION must be SemVer (example: 0.5.0 or v0.5.0)"; \
		exit 1; \
	fi; \
	echo "Stage 1/2 (core tag): $(VERSION_TAG)"; \
	if git rev-parse -q --verify "refs/tags/$(VERSION_TAG)" >/dev/null; then \
		echo "  local:  conflict (exists)"; \
	else \
		echo "  local:  ok"; \
	fi; \
	if git ls-remote --exit-code --tags origin "refs/tags/$(VERSION_TAG)" >/dev/null 2>&1; then \
		echo "  remote: conflict (exists)"; \
	else \
		echo "  remote: ok"; \
	fi; \
	echo "Stage 1/2 also syncs extra deps to $(VERSION_TAG)"; \
	for mod in $(EXTRA_MODULES); do \
		current="$$(cd "$$mod" && go list -m -f '{{.Version}}' github.com/akfaiz/migris)"; \
		if [ "$$current" = "$(VERSION_TAG)" ]; then \
			echo "  [ok]   $$mod -> $$current"; \
		else \
			echo "  [diff] $$mod -> $$current (expected $(VERSION_TAG))"; \
		fi; \
	done; \
	echo "Stage 2/2 (extra tags):"; \
	for tag in extra/migriscli/$(VERSION_TAG) extra/migriscobra/$(VERSION_TAG); do \
		local_status="ok"; \
		remote_status="ok"; \
		if git rev-parse -q --verify "refs/tags/$$tag" >/dev/null; then \
			local_status="conflict"; \
		fi; \
		if git ls-remote --exit-code --tags origin "refs/tags/$$tag" >/dev/null 2>&1; then \
			remote_status="conflict"; \
		fi; \
		echo "  $$tag (local=$$local_status, remote=$$remote_status)"; \
	done; \
	echo "Flow:"; \
	echo "  1) make release-core VERSION=$(VERSION_TAG)"; \
	echo "  2) commit + push updated extra/go.mod and go.sum"; \
	echo "  3) make release-extra VERSION=$(VERSION_TAG)"

.PHONY: delete-tag
delete-tag: # Delete a tag locally and from origin (TAG=v0.5.0)
	@set -e; \
	if [ -z "$(TAG)" ]; then \
		echo "TAG is required. Usage: make delete-tag TAG=v0.5.0"; \
		exit 1; \
	fi; \
	local_exists=false; \
	remote_exists=false; \
	if git rev-parse -q --verify "refs/tags/$(TAG)" >/dev/null; then \
		local_exists=true; \
	fi; \
	if git ls-remote --exit-code --tags origin "refs/tags/$(TAG)" >/dev/null 2>&1; then \
		remote_exists=true; \
	fi; \
	if [ "$$local_exists" = "false" ] && [ "$$remote_exists" = "false" ]; then \
		echo "Tag not found locally or on origin: $(TAG)"; \
		exit 1; \
	fi; \
	if [ "$$local_exists" = "true" ]; then \
		echo "Deleting local tag: $(TAG)"; \
		git tag -d "$(TAG)"; \
	fi; \
	if [ "$$remote_exists" = "true" ]; then \
		echo "Deleting remote tag: $(TAG)"; \
		git push origin ":refs/tags/$(TAG)"; \
	fi

.PHONY: help
help:
	@printf "$(BLUE)Available commands$(NC)\n\n"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "fmt" "Format code"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "lint" "Lint code"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "lint-fix" "Fix lint issues"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "install" "Install dependencies"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "tidy" "Run go mod tidy in all modules"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "install-tools" "Install development tools"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "test" "Run tests"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "testcov" "Run tests with coverage"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "testcov-html" "Generate coverage HTML report"
	@printf "\n$(BLUE)Release flow$(NC)\n\n"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "sync-extra-deps" "Sync extra module dep to root release version"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "release-core-preflight" "Validate root tag release prerequisites"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "release-core-publish" "Create and push root module tag"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "release-extra-preflight" "Validate extra tag release prerequisites"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "release-extra-publish" "Create and push extra module tags"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "release-core" "Stage 1: publish root tag and sync extra deps"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "release-extra" "Stage 2: publish extra module tags"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "release-dry-run" "Dry-run for staged release flow"
	@printf "$(YELLOW)%-24s$(NC) %s\n" "delete-tag TAG=v0.5.0" "Delete local/remote tag by name"
	@printf "\n$(BLUE)Examples$(NC)\n\n"
	@printf "  make release-core VERSION=0.5.0\n"
	@printf "  make release-extra VERSION=0.5.0\n"
	@printf "  make release-dry-run VERSION=v0.5.0\n"
	@printf "\n$(YELLOW)%-24s$(NC) %s\n" "help" "Show this help message"
