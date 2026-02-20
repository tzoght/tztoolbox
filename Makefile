# tztoolbox — Cursor goodies (commands, rules, skills)
# Run: make help | make install

SHELL      := /bin/sh
CURSOR_HOME := $(HOME)/.cursor
REPO_CURSOR := .cursor

SHELLCHECK ?= shellcheck
SHFMT      ?= shfmt

SHELL_SCRIPTS := install.sh

.DEFAULT_GOAL := help

.PHONY: help install fmt lint test test-unit test-integration build ci clean check

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------
help:
	@echo "tztoolbox — Cursor commands, rules, and skills"
	@echo ""
	@echo "Targets:"
	@echo "  make install          Copy commands, rules, and skills into $(CURSOR_HOME)"
	@echo "  make fmt              Auto-format shell scripts (shfmt)"
	@echo "  make lint             Lint shell scripts (shellcheck)"
	@echo "  make test             Run all tests"
	@echo "  make test-unit        Run unit tests (script syntax check)"
	@echo "  make test-integration Run integration tests (dry-run install)"
	@echo "  make build            No-op (nothing to compile)"
	@echo "  make ci               Run full CI pipeline: fmt, lint, test, build"
	@echo "  make clean            Remove generated artifacts"
	@echo "  make check            Alias for ci"
	@echo "  make help             Show this help"

# ---------------------------------------------------------------------------
# Install
# ---------------------------------------------------------------------------
# Always overwrites existing files so updates from git pull are picked up.
install:
	@mkdir -p $(CURSOR_HOME)/commands $(CURSOR_HOME)/rules $(CURSOR_HOME)/skills
	@for f in $(REPO_CURSOR)/commands/*.md; do \
		if [ -f "$$f" ]; then \
			cp "$$f" "$(CURSOR_HOME)/commands/$$(basename "$$f")"; \
		fi; \
	done
	@for f in $(REPO_CURSOR)/rules/*; do \
		if [ -f "$$f" ]; then \
			cp "$$f" "$(CURSOR_HOME)/rules/$$(basename "$$f")"; \
		fi; \
	done
	@for d in $(REPO_CURSOR)/skills/*/; do \
		if [ -d "$$d" ]; then \
			cp -r "$${d%/}" "$(CURSOR_HOME)/skills/"; \
		fi; \
	done
	@echo "Installed commands -> $(CURSOR_HOME)/commands/"
	@echo "Installed rules    -> $(CURSOR_HOME)/rules/"
	@echo "Installed skills   -> $(CURSOR_HOME)/skills/"

# ---------------------------------------------------------------------------
# Format
# ---------------------------------------------------------------------------
fmt:
	@if command -v $(SHFMT) >/dev/null 2>&1; then \
		echo "Formatting shell scripts..."; \
		$(SHFMT) -w -i 2 -ci $(SHELL_SCRIPTS); \
	else \
		echo "shfmt not found — skipping format (install: https://github.com/mvdan/sh)"; \
	fi

# ---------------------------------------------------------------------------
# Lint
# ---------------------------------------------------------------------------
lint:
	@if command -v $(SHELLCHECK) >/dev/null 2>&1; then \
		echo "Linting shell scripts..."; \
		$(SHELLCHECK) $(SHELL_SCRIPTS); \
	else \
		echo "shellcheck not found — skipping lint (install: https://github.com/koalaman/shellcheck)"; \
	fi

# ---------------------------------------------------------------------------
# Test
# ---------------------------------------------------------------------------
test: test-unit test-integration

test-unit:
	@echo "Running unit tests (syntax check)..."
	@for f in $(SHELL_SCRIPTS); do \
		sh -n "$$f" && echo "  $$f — OK"; \
	done

test-integration:
	@echo "Running integration tests (dry-run install)..."
	@tmpdir=$$(mktemp -d) && \
		HOME="$$tmpdir" sh install.sh && \
		echo "  Verifying installed files..." && \
		test -d "$$tmpdir/.cursor/commands" && echo "  commands/ — OK" && \
		test -d "$$tmpdir/.cursor/rules"    && echo "  rules/    — OK" && \
		rm -rf "$$tmpdir" && \
		echo "  Integration test passed."

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------
build:
	@echo "Nothing to compile — content-only project."

# ---------------------------------------------------------------------------
# CI
# ---------------------------------------------------------------------------
ci: fmt lint test build
	@echo ""
	@echo "CI passed."

check: ci

# ---------------------------------------------------------------------------
# Clean
# ---------------------------------------------------------------------------
clean:
	@echo "Nothing to clean."
