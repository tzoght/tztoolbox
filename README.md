# tztoolbox

[![CI](https://github.com/tzoght/tztoolbox/actions/workflows/ci.yml/badge.svg)](https://github.com/tzoght/tztoolbox/actions/workflows/ci.yml)
[![Publish](https://github.com/tzoght/tztoolbox/actions/workflows/publish.yml/badge.svg)](https://github.com/tzoght/tztoolbox/actions/workflows/publish.yml)
![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/tzoght/tztoolbox?utm_source=oss&utm_medium=github&utm_campaign=tzoght%2Ftztoolbox&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)

One repository to manage **commands**, **rules**, and **skills** across three AI tools:

- [Cursor](https://cursor.com) (`~/.cursor`)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (`~/.claude`)
- [OpenAI Codex CLI](https://openai.com/codex/) (`~/.codex`)

Author each artifact once under `shared/`. The native Go CLI, **`tzcli`**, renders it into each tool's native layout and installs it into your home directory.

## At a glance

```
shared/
  commands/       slash-command playbooks (Markdown)
  rules/          always-on guidance (Markdown)
  skills/<name>/  Agent Skills standard layout (SKILL.md + references/)
overrides/
  cursor/rules/   per-rule frontmatter for .mdc files
  claude/         per-tool disabled.yaml + memory snippets
  codex/          per-tool disabled.yaml + AGENTS.md snippets
.cursor/ .claude/ .codex/   generated trees, committed for in-repo UX
cmd/tzcli/        Go entrypoint
internal/         render, installer, doctor, authoring, diffsearch, …
```

## Quick install

```bash
# Default: install for whichever tools are detected on this machine
curl -fsSL https://raw.githubusercontent.com/tzoght/tztoolbox/main/install.sh | sh

# Install for one tool only
curl -fsSL https://raw.githubusercontent.com/tzoght/tztoolbox/main/install.sh | sh -s -- --only cursor
```

The bootstrap downloads the latest `tzcli` binary (or builds from source if Go is available) into `~/.local/bin/tzcli` and runs `tzcli install`. Re-run anytime to update.

## Using inside this repo

Just open the repo. Cursor, Claude Code, and Codex CLI each read their respective `.cursor/`, `.claude/`, and `.codex/` directories at the workspace root, so all three tools work out of the box without running anything.

To regenerate those trees after editing `shared/` or `overrides/`:

```bash
make sync   # or: tzcli sync
```

## `tzcli` reference

| Command | Purpose |
|---|---|
| `tzcli` (no args) | On a TTY, drops into the interactive shell (alias for `tzcli shell`). On a pipe / CI, prints help. |
| `tzcli shell` | Launches the resident, Bubbletea-powered TUI (status bar, command palette, history, tab completion). |
| `tzcli sync [--check] [--only <tool>]` | Render `shared/` + `overrides/` into the in-repo native trees. `--check` exits non-zero on drift. |
| `tzcli install [--only <tool>] [--no-prune]` | Copy the in-repo native trees into `~/.<tool>`. Auto-detects which tools are installed. |
| `tzcli installed [--tool=<t>] [--kind=<k>] [--managed]` | List artifacts deployed under `~/.<tool>` side-by-side with `<repo>/.<tool>` (status: `in_sync`, `only_in_repo`, `only_in_home`, `changed`). |
| `tzcli doctor` | Print detected tools, versions, drift status, helper presence (`gh`, `op`, `git`, `go`). |
| `tzcli update [--remote=origin] [--force]` | Fast-forward `git pull`. Refuses on a dirty tree unless `--force`. |
| `tzcli add <command\|rule\|skill> <name>` | Scaffold a new artifact under `shared/`, then run `sync`. |
| `tzcli list [commands\|rules\|skills]` | Tabular listing of artifacts. |
| `tzcli validate` | Check SKILL.md frontmatter, command/rule H1s, and scan for committed secrets. |
| `tzcli enable <name> --tool=<t> --kind=<k>` | Re-enable an artifact for one tool. |
| `tzcli disable <name> --tool=<t> --kind=<k>` | Drop an artifact from one tool's render via `overrides/<tool>/disabled.yaml`. |
| `tzcli search <pattern> [--in commands,rules,skills] [-i]` | Regex search across `shared/`. |

Global flags: `--repo`, `--verbose`, `--no-color`, `--json`.

Output is colorized and rendered as bordered tables when stdout is a TTY. In pipes or CI it falls back to plain text (`tabwriter` columns) automatically. Set `NO_COLOR=1` or pass `--no-color` to force plain output even on a TTY. `--json` always emits machine-readable output and bypasses the renderer entirely.

## Interactive shell

Run `tzcli` (no args) on a terminal — or `tzcli shell` from anywhere — to drop into the resident TUI:

- A status bar across the top shows version, repo path, current branch, detected tools (cursor / claude / codex), and the current drift count.
- A scrollable viewport shows command output with the same colors and bordered tables that the standalone subcommands use.
- A line-editor prompt at the bottom accepts any `tzcli` subcommand (e.g. `sync --check`, `installed --tool cursor`, `doctor`) plus the built-ins `help`, `clear`, `version`, and `quit`.
- The startup screen prints a menu of every subcommand with a single-letter mnemonic (e.g. `[s] sync`, `[d] doctor`, `[a] add`). Type the letter alone (or with arguments — `s --check`) and the REPL expands it to the full command. Mnemonics live only in the REPL; the standalone CLI surface stays free of single-letter aliases.
- For per-subcommand help, type either `<name> --help` (e.g. `add --help`) or `help <name>`. When a command fails (e.g. missing required args), the error footer surfaces a hint pointing at both forms.

Default key bindings:

| Key | Action |
|---|---|
| `Enter` | Run the current line. |
| `Tab` | Autocomplete subcommands and flags by walking the Cobra tree. |
| `Up` / `Down` | Cycle through command history (persisted across sessions). |
| `Ctrl+K` | Open the command palette to fuzzy-pick a subcommand. |
| `Ctrl+L` | Clear the output viewport. |
| `Ctrl+C` | Cancel the current input line. |
| `Ctrl+D` / `quit` | Exit the shell. |
| `?` | Toggle the full key-binding help. |

History is stored at `${TZCLI_HISTORY_FILE}` if set, otherwise `${XDG_CONFIG_HOME:-~/.config}/tzcli/history`. Each REPL invocation builds a fresh Cobra command tree so flag state never leaks between commands.

## Adding goodies

| Kind | Source location | Cursor | Claude | Codex |
|---|---|---|---|---|
| Command | `shared/commands/<name>.md` | `.cursor/commands/<name>.md` | `.claude/commands/<name>.md` | `.codex/prompts/<name>.md` |
| Rule | `shared/rules/<name>.md` (+ `overrides/cursor/rules/<name>.frontmatter.yaml` for Cursor scoping) | `.cursor/rules/<name>.mdc` | concatenated into `.claude/CLAUDE.md` | concatenated into `.codex/AGENTS.md` |
| Skill | `shared/skills/<name>/SKILL.md` (+ `references/`, `scripts/`, etc.) | `.cursor/skills/<name>/` | `.claude/skills/<name>/` | `.codex/skills/<name>/` |

Use the scaffolders to start from the right shape:

```bash
tzcli add command tz-my-thing
tzcli add rule    my-team-rule
tzcli add skill   my-skill
```

After editing, run `make ci` to format, lint, test, and build before opening a PR.

## Development

```bash
make build              # go build -o bin/tzcli
make sync               # render shared/ -> .cursor, .claude, .codex
make test               # go test (unit + integration)
make ci                 # fmt -> lint -> test -> build
```

## Code reviews

Pull requests are automatically reviewed by [CodeRabbit](https://coderabbit.ai). The configuration lives in [`.coderabbit.yaml`](.coderabbit.yaml).

## License

See [LICENSE](LICENSE).
