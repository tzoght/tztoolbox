# tztoolbox

[![CI](https://github.com/tzoght/tztoolbox/actions/workflows/ci.yml/badge.svg)](https://github.com/tzoght/tztoolbox/actions/workflows/ci.yml)
[![Publish](https://github.com/tzoght/tztoolbox/actions/workflows/publish.yml/badge.svg)](https://github.com/tzoght/tztoolbox/actions/workflows/publish.yml)
[![CodeRabbit](https://img.shields.io/badge/CodeRabbit-AI%20Reviews-blue?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0Ij48cGF0aCBkPSJNMTIgMkM2LjQ4IDIgMiA2LjQ4IDIgMTJzNC40OCAxMCAxMCAxMCAxMC00LjQ4IDEwLTEwUzE3LjUyIDIgMTIgMnoiIGZpbGw9IiNmZmYiLz48L3N2Zz4=)](https://coderabbit.ai)

Useful prompts, shells, and tools collected over time. This repo holds Cursor **commands**, **rules**, and **skills** so you can reuse them everywhere.

## Using in this repo

Open this repo in Cursor:

- **Commands** — Type `/` in the chat; commands from `.cursor/commands/` appear in the palette.
- **Rules & skills** — Loaded automatically from `.cursor/rules/` and `.cursor/skills/`.

No install step needed when you’re working inside this repo.

## Quick install (no clone needed)

```bash
curl -fsSL https://raw.githubusercontent.com/tzoght/tztoolbox/main/install.sh | sh
```

This clones the repo into a temp directory, copies all commands, rules, and skills into `~/.cursor/`, and cleans up. Re-run anytime to update.

## Using in other projects

If you already have a local clone, you can install with Make instead:

```bash
make install
```

This copies:

- `.cursor/commands/` → `~/.cursor/commands/`
- `.cursor/rules/` → `~/.cursor/rules/`
- `.cursor/skills/*` → `~/.cursor/skills/` (user skills; Cursor reserves `skills-cursor/` for built-ins)

Re-run after `git pull` to update your global copy. If Rules or Skills don’t show in **Cursor Settings → Rules, Skills, Subagents**, restart Cursor once after installing.

**Other targets:** run `make` or `make help` to list targets.

## Adding goodies

| Type    | Where to add                         | Format |
|---------|--------------------------------------|--------|
| Command | `.cursor/commands/`                  | One `.md` per command; first heading = name in `/` |
| Rule    | `.cursor/rules/`                     | One `.md` or `.mdc` per rule (use frontmatter for scope) |
| Skill   | `.cursor/skills/<name>/`             | One directory per skill with a `SKILL.md` inside |

Then commit, push, and run `make install` again to refresh the global copy.

## Code reviews

Pull requests are automatically reviewed by [CodeRabbit](https://coderabbit.ai). The configuration lives in [`.coderabbit.yaml`](.coderabbit.yaml). CodeRabbit needs to be installed as a GitHub App on the repository — see the [setup guide](https://docs.coderabbit.ai/platforms/github-com) if it isn't already.

## License

See [LICENSE](LICENSE).
