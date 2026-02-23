# Create Branch

When the user invokes this command, run the following workflow to create a new feature branch from the latest upstream HEAD.

## 0. Pre-flight checks

Run all three checks **in parallel** (they are independent):

1. `git rev-parse --is-inside-work-tree 2>/dev/null` — if not a git repo, stop: **"Not a git repository."**
2. `git status --porcelain` — if there are uncommitted changes, warn the user and offer to either **stash** them (`git stash`) or **abort**. Do not proceed with dirty state unless the user explicitly says to carry changes over.
3. `git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@'` — detect the default branch. Fall back to `main` if unset. Remember this as `<default-branch>`.

## 1. Fetch latest

Run `git fetch origin <default-branch>` to update the remote-tracking ref. **Do not checkout or pull the default branch** — the new branch will be created directly from `origin/<default-branch>`, which avoids two working-tree rewrites and is significantly faster on large repos.

## 2. Choose a branch name

Ask the user: **"What is this branch for?"**
- Accept a short description (e.g. "add login page", "fix header bug", "update deps").
- Derive a branch name using a conventional prefix based on the description:
  - `feature/<slug>` — new functionality (default if unclear)
  - `fix/<slug>` — bug fix
  - `chore/<slug>` — maintenance, config, tooling
  - `docs/<slug>` — documentation only
  - `refactor/<slug>` — code restructuring without behavior change
- Slugify the description: lowercase, replace spaces with hyphens, strip special characters, truncate to ~50 chars.
- Present the proposed branch name to the user and ask for confirmation. Let them accept, edit, or replace it.

## 3. Create the branch

1. Check if the branch already exists **in parallel**: locally (`git branch --list <name>`) and remotely (`git branch -r --list origin/<name>` — uses the local cache updated by the fetch in Step 1, avoiding a second network round-trip). If it exists, warn the user and ask whether to switch to it or pick a different name.
2. Run `git checkout -b <name> origin/<default-branch>` to create the new branch directly from the fetched remote HEAD. This is a single operation instead of checkout + pull + branch.
3. If changes were stashed in Step 0, ask the user if they want to restore them (`git stash pop`).

## 4. Summary

Show a recap:

- **Created branch**: `<name>` from `origin/<default-branch>` at `<short-sha>`
- **Current branch**: confirm the user is now on the new branch
- **Stashed changes**: restored / still stashed / none

Keep the flow conversational: confirm naming before creating, and inform the user they can start working.
