# Fix PR Check Failures

**When to use:** The current branch has an open PR (or you expect one) and merge-blocking GitHub checks / Actions are failing.

Find failing checks on that PR, reproduce them locally, fix them, and (optionally) commit and push. Targets GitHub Actions and other status checks, not bot/human review threads (CodeRabbit, human review threads — out of scope for v1).

## Interaction mode

If running with `--force` / non-interactive headless: gather context, print a one-screen plan, execute end-to-end, emit one summary line. Otherwise (REPL): group related questions into one prompt; only confirm before destructive actions; do not narrate intermediate state unless asked.

## 0. Pre-flight (parallel)

1. `git rev-parse --is-inside-work-tree`. Stop if not a git repo.
2. `git branch --show-current`. Stop if empty (detached HEAD); suggest creating/checking out a named branch.
3. `git status --short`. If dirty, warn that fixes will mix with existing changes unless stashed; single ask: stash / proceed anyway / abort.
4. `git fetch origin --prune`.
5. Detect default branch: `git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@'` (fallback `main`). Call it `<default>`.
6. `gh --version 2>/dev/null`. Note whether `gh` is available.

## 1. Resolve the PR

If `gh` works: `gh pr view --json number,url,state,title,statusCheckRollup` for the **current branch**.

- If **no PR**: stop with a short message (“Open or create a PR for this branch first”). Point to [tz-changes-publish](tz-changes-publish.md) and, if `owner/repo` is known from `git remote get-url origin`, a compare URL: `https://<host>/<owner>/<repo>/compare/<default>...<current>?expand=1`.
- If the PR is **closed/merged**: single ask whether to continue with local CI only.

If `gh` fails (auth, EMU, network): derive `owner/repo` and `<host>` from `git remote get-url origin`; print the compare URL above (or `https://<host>/<owner>/<repo>/pulls` if the branch name is unknown). Skip Step 2 and continue with **local-only** failure finding (Step 3).

## 2. Remote failure summary (when `gh` works)

1. `gh pr checks` (or check fields from Step 1) — list failing / pending / success. **Do not** paste huge logs; summarize root cause.
2. If checks are **pending**, optionally `gh pr checks --watch` with a short timeout (2–3 minutes), or tell the user to re-run this command when CI finishes.
3. For each **failed** Actions run: `gh run list --branch <current> --status failure --limit 5`, then `gh run view <run-id> --log-failed` (pipe through `tail -200` if needed). Summarize workflow, step, file, and error line.

## 3. Local reproduction

Prefer the repo’s standard local CI gate (call it `<ci-cmd>`):

1. If the repo root has a **Makefile** with a `ci` target: `<ci-cmd>` = `make ci` (matches typical `.github/workflows/` jobs in Makefile-based repos).
2. Else if Step 2 named a workflow command: mirror that command from `.github/workflows/<file>.yml`.
3. Else: `make check`, or `make lint && make test && make build` (only targets that exist), or `package.json` scripts `ci` / `check` / `lint`+`test`.

Run `<ci-cmd>` from the repo root. If it passes locally but GitHub still shows red, confirm `git log origin/<current>..HEAD` is empty (fully pushed), then compare local `HEAD` SHA with the failing workflow run’s commit SHA and explain likely drift (unpushed commits, wrong branch, workflow-only environment).

## 4. Fix loop

1. Map the failure to the smallest matching step in `<ci-cmd>` (for Makefile-based repos: usually `fmt`, `lint`, `test`, or `build`).
2. Apply minimal fixes; re-run the failing step, then the full `<ci-cmd>` until green.
3. Repeat until `<ci-cmd>` passes.

If failure is **only** on GitHub (runner image, missing apt package, env only in Actions), edit this repo’s `.github/workflows/` and any install/scripts the repo owns. Do not change org-level settings unless asked.

## 5. Ship fixes (optional, conversational)

Show `git diff --stat` and a suggested conventional commit message (e.g. `fix(ci): ...`). Single `[y/edit/n]` before `git commit` and `git push`. In headless `--force`: accept defaults. Never force-push unless explicitly requested.

Do **not** call `gh pr create` or other `gh` write APIs if auth/EMU blocks them; provide the PR URL from Step 1 for manual refresh after push.

Emit one line: `<ci-cmd> green; <committed|uncommitted>; <pushed|not pushed>; PR <url|open manually>`.
