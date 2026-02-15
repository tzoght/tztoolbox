# tztoolbox — Cursor goodies (commands, rules, skills)
# Run: make help | make install

SHELL := /bin/sh
CURSOR_HOME := $(HOME)/.cursor
REPO_CURSOR := .cursor

.PHONY: help install

help:
	@echo "tztoolbox — Cursor commands, rules, and skills"
	@echo ""
	@echo "Targets:"
	@echo "  make install   Copy .cursor/commands, rules, and skills into $(CURSOR_HOME)"
	@echo "  make help      Show this help"
	@echo ""
	@echo "After install, goodies are available in all Cursor projects. Re-run after git pull to update."

# User skills go in ~/.cursor/skills/ (skills-cursor is reserved for Cursor built-ins)
# Idempotent: skip copying if command/rule/skill already exists at destination.
install:
	@mkdir -p $(CURSOR_HOME)/commands $(CURSOR_HOME)/rules $(CURSOR_HOME)/skills
	@for f in $(REPO_CURSOR)/commands/*.md; do \
		if [ -f "$$f" ]; then \
			dest="$(CURSOR_HOME)/commands/$$(basename "$$f")"; \
			if [ ! -f "$$dest" ]; then cp "$$f" "$$dest"; fi; \
		fi; \
	done
	@for f in $(REPO_CURSOR)/rules/*; do \
		if [ -f "$$f" ]; then \
			dest="$(CURSOR_HOME)/rules/$$(basename "$$f")"; \
			if [ ! -f "$$dest" ]; then cp "$$f" "$$dest"; fi; \
		fi; \
	done
	@for d in $(REPO_CURSOR)/skills/*/; do \
		if [ -d "$$d" ]; then \
			name=$$(basename "$${d%/}"); \
			if [ ! -d "$(CURSOR_HOME)/skills/$$name" ]; then cp -r "$${d%/}" "$(CURSOR_HOME)/skills/"; fi; \
		fi; \
	done
	@echo "Installed commands -> $(CURSOR_HOME)/commands/"
	@echo "Installed rules    -> $(CURSOR_HOME)/rules/"
	@echo "Installed skills   -> $(CURSOR_HOME)/skills/"

# Default target
.DEFAULT_GOAL := help
