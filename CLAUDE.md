# tztoolbox — Claude Code orientation

This repo is a multi-tool toolbox: it manages **commands**, **rules**, and **skills**
for Cursor, Claude Code, and OpenAI Codex CLI from one source of truth.

## Where things live

- `shared/commands/<name>.md` — slash-command playbooks (one H1 = command title).
- `shared/rules/<name>.md` — always-on guidance (no frontmatter; concatenated into `.claude/CLAUDE.md` for you).
- `shared/skills/<name>/SKILL.md` — Agent Skills (frontmatter: `name`, `description` ≤200 chars, third-person).
- `overrides/<tool>/` — per-tool tweaks (Cursor `.mdc` frontmatter, `disabled.yaml`, …).
- `.cursor/`, `.claude/`, `.codex/` — generated trees, committed for in-repo UX. Edit `shared/`/`overrides/` instead.

## How to make changes

1. Edit `shared/...` (or scaffold via `tzcli add ...`).
2. Run `tzcli sync` (or `make sync`) to regenerate the native trees.
3. Run `make ci` before committing — it formats, lints (golangci-lint + shellcheck), tests, and builds.
4. Open a feature branch (workspace rule: never commit directly to `main`).

## tzcli at a glance

`tzcli` is the native Go CLI under `cmd/tzcli`. Subcommands worth knowing:
`sync`, `install`, `doctor`, `update`, `add`, `list`, `validate`,
`enable`/`disable`, `search`, `shell`. Run `tzcli --help` for details.

## Don't do

- Don't edit files under `.cursor/`, `.claude/`, or `.codex/` directly — they are regenerated.
- Don't commit secrets. The `1password-mcp` skill is placeholder-only by design.
- Don't add Cursor `.mdc` frontmatter inside `shared/rules/` — it lives in `overrides/cursor/rules/<name>.frontmatter.yaml`.
