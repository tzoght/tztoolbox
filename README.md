# tztoolbox

Useful prompts, shells, and tools collected over time. This repo holds Cursor **commands**, **rules**, and **skills** so you can reuse them everywhere.

## Using in this repo

Open this repo in Cursor:

- **Commands** — Type `/` in the chat; commands from `.cursor/commands/` appear in the palette.
- **Rules & skills** — Loaded automatically from `.cursor/rules/` and `.cursor/skills/`.

No install step needed when you’re working inside this repo.

## Using in other projects

To get the same goodies in every Cursor project (without opening this repo):

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

## License

See [LICENSE](LICENSE).
