# Delete Branch

Safely delete a feature branch locally and remotely. Accepts a branch name as the first argument; otherwise prompts for one.

## Interaction mode

If running with `--force` / non-interactive headless: gather context, print a one-screen plan, execute end-to-end, emit one summary line. Otherwise (REPL): group related questions into one prompt; only confirm before destructive actions; do not narrate intermediate state unless asked.

## 0. Pre-flight (parallel)

1. `git rev-parse --is-inside-work-tree`. Stop if not a git repo.
2. `git branch --show-current`. Remember as `<current>`.
3. `git fetch --prune origin`. Sync remote-tracking refs.
4. `git branch` and `git branch -r`. Local + remote branch lists.
5. `gh --version`. Note availability for the PR query below.

## 1. Pick the branch

If the user supplied a branch name as the first argument, use it directly (skip the prompt). Otherwise present the branches from Step 0 as a numbered list, excluding `main`, `master`, `develop`, `staging`, `release/*`, `production`. Refuse protected branches if picked.

## 2. Safety checks (parallel reads)

- **Merged?** Run in parallel: `git branch --merged <default>`, `git branch -r --merged origin/<default>`, and (if `gh` available) `gh pr list --head <branch> --state all --json number,url,state` (single call covers open + merged including squash/rebase). Treat the branch as merged if any indicator says so.
- **Open PR?** Filter the same `gh pr list` result for `open` entries; capture the URL if present.
- **Current-branch conflict?** If `<current> == <branch>`, plan to switch to `<default>` first.

## 3. Single confirmation gate

Show one line and wait for `[y/n]` (in headless `--force`, treat as yes):

`Delete '<branch>' [merged via PR #N | unmerged: K commits would be lost] locally and on origin [open PR #M still attached]?`

## 4. Execute

1. If on the branch, `git checkout <default>`.
2. Local: `git branch -d <branch>` (or `-D <branch>` if unmerged and the user confirmed).
3. Remote: `git push origin --delete <branch>` (skip silently if remote ref is absent).

## 5. Done

Emit one line: `Deleted <branch>: local <ok|skipped>, remote <ok|skipped>; now on <default>`.
