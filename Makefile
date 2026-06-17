# artiworks Makefile
.DEFAULT_GOAL := help

BINARY  := artiworks
GO      := go
GOFLAGS := -v
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# ── Development ──────────────────────────────────────────

.PHONY: build dev test clean schema tools install-skills lint lint-fix fmt vet audit outdated install deps-update help

build: ## Build CLI binary
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY) ./cmd/artiworks
	@echo "✅ $(BINARY) $(VERSION)"

dev: ## Build and run
	@$(GO) run ./cmd/artiworks

test: ## Run all tests with race detector + coverage
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html 2>/dev/null || true

clean: ## Remove build artifacts (cross-platform)
	@$(GO) run scripts/clean.go && echo "✅ cleaned"

schema: ## Regenerate JSON Schema from Go types
	@$(GO) generate ./pkg/artiworks/config
	@echo "✅ schema.json"

# ── Tools ────────────────────────────────────────────────

GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null)
GOVULNCHECK   := $(shell command -v govulncheck 2>/dev/null)

tools: ## Install development tools (golangci-lint, govulncheck)
ifndef GOLANGCI_LINT
	@echo "Installing golangci-lint..."
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
endif
ifndef GOVULNCHECK
	@echo "Installing govulncheck..."
	@$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
endif
	@echo "✅ tools ready"

install-skills: ## Install Go agent skills (requires npx)
	@npx skills add https://github.com/samber/cc-skills-golang --all

# ── Quality ──────────────────────────────────────────────

lint: ## Run golangci-lint
	@golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	@golangci-lint run --fix ./...

fmt: ## Format and vet
	@$(GO) fmt ./...
	@$(GO) vet ./...

audit: ## Audit dependencies for vulnerabilities
	@govulncheck ./...

outdated: ## Check for outdated dependencies
	@$(GO) list -u -m -json all 2>/dev/null | grep -E '"Path"|"Version"' | head -30 || echo "run: go list -u -m all"

# ── Dependencies ─────────────────────────────────────────

install: ## Download and tidy dependencies
	@$(GO) mod download
	@$(GO) mod tidy

deps-update: ## Update dependencies (patch only)
	@$(GO) get -u=patch ./...
	@$(GO) mod tidy

# ── Help ─────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}'

%::
	@echo "Unknown target '$@'. Run 'make help'."
	@exit 1
