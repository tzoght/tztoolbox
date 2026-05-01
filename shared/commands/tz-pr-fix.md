# Fix PR Check Failures

Find failing checks on the open PR for the current branch, reproduce them locally, fix them, and (optionally) commit and push. Targets GitHub Actions and other status checks, not bot/human review threads.

## Interaction mode

If running with `--force` / non-interactive headless: gather context, print a one-screen plan, execute end-to-end, emit one summary line. Otherwise (REPL): group related questions into one prompt; only confirm before destructive actions; do not narrate intermediate state unless asked.

## 0. Pre-flight (parallel)

1. `git rev-parse --is-inside-work-tree`. Stop if not a git repo.
2. `git branch --show-current`. Stop if empty (detached HEAD).
3. `git status --short`. If dirty, single ask: stash / proceed anyway / abort.
4. `git fetch origin --prune`.
5. Detect default branch: `git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@'` (fallback `main`).
6. `gh --version`. Note availability.

## 1. Resolve the PR

If `gh` works: `gh pr view --json number,url,state,title,statusCheckRollup`. If no PR for this branch, stop and point to [tz-changes-publish](tz-changes-publish.md). If the PR is closed/merged, single ask: continue with local CI only?

If `gh` fails (auth, EMU, network): derive `owner/repo` from `git remote get-url origin` (only now, not in Step 0); print `https://<host>/<owner>/<repo>/pulls`; skip Step 2 and proceed to Step 3.

## 2. Snapshot remote failures (when `gh` works)

1. `gh pr checks` (snapshot only; **do not** `--watch`). If checks are still pending, tell the user to re-run after CI finishes and stop here.
2. For each failed Actions workflow: `gh run list --branch <current> --status failure --limit 5` for the latest run id, then `gh run view <run-id> --log-failed | tail -200`. Summarize root cause (file, step, error line); never paste full logs into chat.

## 3. Local reproduction

`make ci`. If it passes locally but GitHub still shows red, confirm `git log origin/<current>..HEAD` is empty (i.e. fully pushed), then compare local HEAD SHA with the failing workflow's commit SHA and explain likely drift.

## 4. Fix loop (targeted retries)

1. Map the first failure to its `make` target: `fmt`, `lint`, `test`, or `build`.
2. Make the minimal edit and re-run **only that target** (e.g. `make lint` after a lint fix).
3. Repeat until that target passes.
4. When the individual targets are green, run **one** full `make ci` as the final gate.

If only the GitHub-side workflow fails (tool install path, missing workflow package), edit [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) in this repo. Do not change org-level settings unless asked.

## 5. Ship (single gate)

Show `git diff --stat` and a suggested conventional commit message (e.g. `fix(ci): ...`). Single `[y/edit/n]` to commit + push. In headless `--force`: accept defaults. Never force-push unless explicitly requested.

## 6. Done

Emit one line: `Fixed <N> checks; pushed <sha> to origin/<current>; PR <url> will refresh on the next CI run`.
