# tztoolbox — Cursor goodies (commands, rules, skills)
# Run: make help | make install

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
install:
	@mkdir -p $(CURSOR_HOME)/commands $(CURSOR_HOME)/rules $(CURSOR_HOME)/skills
	@cp $(REPO_CURSOR)/commands/*.md $(CURSOR_HOME)/commands/ 2>/dev/null || true
	@cp $(REPO_CURSOR)/rules/* $(CURSOR_HOME)/rules/ 2>/dev/null || true
	@for d in $(REPO_CURSOR)/skills/*/; do \
		[ -d "$$d" ] && cp -r "$${d%/}" "$(CURSOR_HOME)/skills/"; \
	done
	@echo "Installed commands -> $(CURSOR_HOME)/commands/"
	@echo "Installed rules    -> $(CURSOR_HOME)/rules/"
	@echo "Installed skills   -> $(CURSOR_HOME)/skills/"

# Default target
.DEFAULT_GOAL := help
