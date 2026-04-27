# tztoolbox — Cursor / Claude Code / OpenAI Codex artifact toolbox.
# All real work lives in the `tzcli` Go binary; this Makefile is a thin
# wrapper that satisfies workspace SDLC rules and provides language-
# agnostic targets.
#
# Run: make help

SHELL := /bin/sh

GO         ?= go
GOFMT      ?= gofmt
GOLANGCI   ?= golangci-lint
SHELLCHECK ?= shellcheck

BIN_DIR    := bin
TZCLI      := $(BIN_DIR)/tzcli
PKG        := ./...
SHELL_SCRIPTS := install.sh

# Inject version metadata at build time.
VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

.DEFAULT_GOAL := help

.PHONY: help build install sync doctor fmt lint test test-unit test-integration ci check clean dev tzcli agent

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------
help:
	@echo "tztoolbox — multi-tool (Cursor, Claude Code, Codex CLI) artifact toolbox"
	@echo ""
	@echo "Lifecycle:"
	@echo "  make build              Build $(TZCLI)"
	@echo "  make sync               Render shared/ + overrides/ into .cursor, .claude, .codex"
	@echo "  make install            Install rendered trees into ~/.cursor, ~/.claude, ~/.codex"
	@echo "  make doctor             Print environment + drift report"
	@echo "  make agent              Launch Cursor CLI (agent) in this workspace"
	@echo ""
	@echo "Validation:"
	@echo "  make fmt                gofmt -w; goimports if available"
	@echo "  make lint               golangci-lint + shellcheck install.sh + tzcli validate"
	@echo "  make test               Run unit + integration tests"
	@echo "  make test-unit          Unit tests only"
	@echo "  make test-integration   Integration tests (build tag: integration)"
	@echo "  make ci                 fmt -> lint -> test -> build (fail fast)"
	@echo "  make check              Alias for ci"
	@echo "  make clean              Remove $(BIN_DIR) and generated trees"

# ---------------------------------------------------------------------------
# Build / lifecycle
# ---------------------------------------------------------------------------
$(TZCLI): $(shell find cmd internal -name '*.go' 2>/dev/null) go.mod go.sum
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(TZCLI) ./cmd/tzcli

build: $(TZCLI)

tzcli: $(TZCLI)

sync: $(TZCLI)
	$(TZCLI) sync

install: $(TZCLI)
	$(TZCLI) install

doctor: $(TZCLI)
	$(TZCLI) doctor

dev: $(TZCLI)
	$(TZCLI) sync --check

# Launch the Cursor CLI (`agent` / `cursor-agent`) bound to this repo so
# that `tz-*` slash commands and Skills rendered by `tzcli install` are
# discovered. Requires the Cursor CLI to be installed (see `make doctor`)
# and authenticated (`agent login` or `CURSOR_API_KEY`).
agent:
	@if command -v agent >/dev/null 2>&1; then \
		agent --workspace $(CURDIR); \
	elif command -v cursor-agent >/dev/null 2>&1; then \
		cursor-agent --workspace $(CURDIR); \
	else \
		echo "Cursor CLI not found. Install with: curl https://cursor.com/install -fsS | bash"; \
		exit 127; \
	fi

# ---------------------------------------------------------------------------
# Format / lint / test
# ---------------------------------------------------------------------------
fmt:
	$(GOFMT) -w cmd internal
	@if command -v shfmt >/dev/null 2>&1; then \
		shfmt -w -i 2 -ci $(SHELL_SCRIPTS); \
	fi

lint: $(TZCLI)
	@if command -v $(GOLANGCI) >/dev/null 2>&1; then \
		$(GOLANGCI) run; \
	else \
		echo "golangci-lint not installed; running 'go vet' as a fallback"; \
		$(GO) vet $(PKG); \
	fi
	@if command -v $(SHELLCHECK) >/dev/null 2>&1; then \
		$(SHELLCHECK) $(SHELL_SCRIPTS); \
	else \
		echo "shellcheck not installed; skipping shell lint"; \
	fi
	$(TZCLI) validate

test: test-unit test-integration

test-unit:
	$(GO) test -short $(PKG)

test-integration: $(TZCLI)
	$(GO) test -tags=integration $(PKG)

ci: fmt lint test build
	@echo ""
	@echo "CI passed."

check: ci

clean:
	rm -rf $(BIN_DIR)
