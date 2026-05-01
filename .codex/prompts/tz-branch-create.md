# Create Branch

Create a new feature branch from the latest upstream HEAD.

## Interaction mode

If running with `--force` / non-interactive headless: gather context, print a one-screen plan, execute end-to-end, emit one summary line. Otherwise (REPL): group related questions into one prompt; only confirm before destructive actions; do not narrate intermediate state unless asked.

## 0. Pre-flight (parallel)

1. `git rev-parse --is-inside-work-tree`. Stop if not a git repo.
2. `git status --porcelain`. Note dirty state for Step 1.
3. Detect default branch: `git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@'` (fallback `main`). Call it `<default>`.
4. `git fetch origin <default>` (do not checkout `<default>`; create directly from `origin/<default>`).

## 1. Single ask

- Dirty tree: `Working tree has N modified files. What's this branch for? (I'll stash first.)`
- Clean tree: `What's this branch for?`

Slugify the answer; pick a conventional prefix (`feature/`, `fix/`, `chore/`, `docs/`, `refactor/`); show the proposed name; let the user accept, edit, or replace it.

## 2. Create

1. Verify the name is free locally (`git branch --list <name>`) and remotely (`git branch -r --list origin/<name>`, served from the cache fetched in Step 0). If taken, ask whether to switch or rename.
2. Stash if dirty, then `git checkout -b <name> origin/<default>`.
3. Restore stash: auto-pop in headless; one-line REPL ask (`Restore stashed changes? [Y/n]`).

## 3. Done

Emit one line: `Branch <name> created from origin/<default> at <sha>; stash <restored|kept|none>`.
