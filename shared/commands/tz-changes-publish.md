# Publish Changes

Stage, commit, push, and (optionally) open a PR for the current branch.

## Interaction mode

If running with `--force` / non-interactive headless: gather context, print a one-screen plan, execute end-to-end, emit one summary line. Otherwise (REPL): group related questions into one prompt; only confirm before destructive actions; do not narrate intermediate state unless asked.

## 0. Pre-flight (parallel)

1. `git status --short`. If clean, stop: **"Nothing to stage."** Keep this output for Step 1.
2. `git branch --show-current`. If `main`, `master`, or the repo default, warn and require explicit confirmation (or a switch) before proceeding.
3. `gh --version`. Remember availability for Step 2.

## 1. Single plan-gate

Build a one-screen plan from the captured `git status` and present it for a single `[y/edit/n]` decision. In headless `--force`: skip the gate, accept defaults.

```
Stage:   <files from Step 0, default "all">
Commit:  "<conventional-style message derived from git diff --stat of the proposed staging set>"
Push:    origin/<branch>   (with -u on first push)
PR:      gh pr create --fill   (or "skip" if disabled / no gh)
```

`edit` lets the user adjust the file selection, the commit message, or skip the PR step.

## 2. Execute

1. `git add <selection>` (or `git add -A` for "all"). Do not echo `git status` again.
2. `git commit -m "<message>"`.
3. `git push` (`-u origin <branch>` on first push).
4. If `gh` available and PR not skipped: in parallel with the push, `gh pr view --json url 2>/dev/null` to detect an existing PR. If one exists, return its URL; otherwise `gh pr create --fill`.
5. If `gh` is unavailable: print the branch name and a suggested PR URL for the user to open manually.

## 3. Done

Emit one line: PR URL if created or found; otherwise `Pushed origin/<branch> at <sha>; open PR manually`.
