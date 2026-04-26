# Fix PR Check Failures

When the user invokes this command, run the following workflow to find failing checks on the **open pull request for the current branch**, reproduce them locally (especially `make ci` when that matches the repo’s GitHub Actions), fix the underlying issues, and optionally commit and push.

**When to use:** The user has a PR open from the current branch and wants CI/check failures investigated and fixed. This command targets **GitHub Actions and other status checks**, not bot or human review comment threads (handle those separately).

## 0. Pre-flight checks

Run **in parallel** wherever independent:

1. `git rev-parse --is-inside-work-tree 2>/dev/null` — if not a git repo, stop: **"Not a git repository."**
2. `git branch --show-current` — if empty (detached HEAD), stop: **"Detached HEAD. Create or checkout a branch, then re-run."**
3. `git status --short` — if the tree is dirty, warn that fixes will mix with uncommitted work unless the user stashes. Offer to stash, proceed anyway, or abort before changing files.
4. `git fetch origin --prune` — sync remote refs so PR and default-branch state are current.
5. Detect the default branch: `git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@'`. Fall back to `main` if unset. Remember as `<default-branch>`.
6. `gh --version 2>/dev/null` — remember whether the GitHub CLI is available.
7. Optionally derive `owner/repo` from `git remote get-url origin` for compare URLs if `gh` is unusable.

If the user aborts due to dirty state, stop.

## 1. Resolve the current PR

**If `gh` is available:**

1. Run `gh pr view --json number,url,state,title,statusCheckRollup 2>/dev/null` for the **current branch** (default behavior of `gh pr view` when run from the repo root on a branch with a PR).
2. If the command fails or returns no PR for this branch, stop: **"No open PR found for this branch."** Point the user to [tz-changes-publish](tz-changes-publish.md) and suggest opening a PR or confirm they are on the correct branch.
3. If the PR is not **open** (e.g. closed/merged), warn and ask whether to continue with local CI only.

**If `gh` fails (auth, Enterprise Managed User, network):**

- Parse `owner/repo` from `git remote get-url origin` when possible.
- Tell the user to open the PR in the browser: `https://github.com/<owner>/<repo>/pulls` (or their host if not `github.com`).
- Skip Steps 2’s remote check listing and go to **Step 3** for local reproduction, then **Step 4** for fixes.

## 2. Remote failure summary (when `gh` works)

1. Run `gh pr checks` (optionally `gh pr checks <number>` if disambiguation is needed) and classify checks as **pass**, **fail**, or **pending**.
2. For **failed** GitHub Actions workflows on this branch, run `gh run list --branch <current-branch> --status failure --limit 5` and note the latest relevant `run id`.
3. For failures needing detail, run `gh run view <run-id> --log-failed`. **Summarize** root cause (file, step, error line); do not dump entire logs into chat.
4. If checks are **pending**, either wait briefly with `gh pr checks --watch` (bounded wait, e.g. 2–3 minutes) or tell the user to re-run the command after CI finishes.

If every check already passes remotely, still run **Step 3** to confirm local parity.

## 3. Local reproduction

From the repository root:

1. Run `make ci` when a `Makefile` with a `ci` target exists (this repo’s [.github/workflows/ci.yml](../../.github/workflows/ci.yml) runs `make ci`).
2. If `make ci` passes locally but GitHub still shows failures:
   - Confirm the latest commits are **pushed** (`git log origin/<current-branch>..HEAD` should be empty after push).
   - Compare the failing workflow’s commit SHA on GitHub with local `HEAD`.
   - Explain likely drift (unpushed changes, wrong branch, workflow-only differences).

If `make ci` fails, the failure output is the primary signal for **Step 4**.

**Fallback when `Makefile` or `ci` target is missing:** run whatever the project documents as its CI gate (e.g. `make lint && make test`) or read `.github/workflows/*.yml` for the exact commands.

## 4. Fix loop

1. Map failures to fixes: for this repo, most issues come from `fmt`, `lint`, `test`, or `build` stages in the [`Makefile`](../../Makefile). Prefer **minimal, targeted** edits.
2. After each meaningful fix, re-run `make ci` (or the local equivalent) until it succeeds.
3. If only CI on GitHub fails (e.g. tool install path, missing package in workflow), fix [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) or documented setup **in this repo**—do not change org-level settings unless the user asks.

Repeat until local CI is green and, when `gh` works, remote checks match expectations (or the user accepts pushing and waiting for a new run).

## 5. Ship fixes (optional, conversational)

1. Show `git diff --stat` (and summarize substantive changes).
2. Suggest a **conventional** commit message (e.g. `fix(ci): …` or `fix: …`).
3. Ask before **`git commit`** and **`git push`**. Do not force-push unless the user explicitly requests it and understands the impact.

If **`gh pr create`** or other write operations are blocked (e.g. Enterprise Managed User), do not rely on the API to update the PR; provide the PR **URL** from Step 1 and remind the user the push will refresh the PR checks in the browser.

Keep the flow conversational: confirm before commits, pushes, or stash/pop operations.
