# Create Branch

When the user invokes this command, run the following workflow to create a new feature branch from the latest upstream HEAD.

## 0. Pre-flight checks

- Run `git rev-parse --is-inside-work-tree 2>/dev/null`. If not a git repo, stop: **"Not a git repository."**
- Run `git status --porcelain`. If there are uncommitted changes, warn the user and offer to either **stash** them (`git stash`) or **abort**. Do not proceed with dirty state unless the user explicitly says to carry changes over.
- Detect the default branch: run `git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@'`. Fall back to `main` if unset. Remember this as `<default-branch>`.

## 1. Pull latest

1. Run `git checkout <default-branch>` to switch to the default branch.
2. Run `git pull origin <default-branch>` to fetch and merge the latest upstream changes, so the new branch starts from a current HEAD.

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

1. Check if the branch already exists locally (`git branch --list <name>`) or remotely (`git ls-remote --heads origin <name>`). If it does, warn the user and ask whether to switch to it or pick a different name.
2. Run `git checkout -b <name>` to create and switch to the new branch.
3. If changes were stashed in Step 0, ask the user if they want to restore them (`git stash pop`).

## 4. Summary

Show a recap:

- **Created branch**: `<name>` from `<default-branch>` at `<short-sha>`
- **Current branch**: confirm the user is now on the new branch
- **Stashed changes**: restored / still stashed / none

Keep the flow conversational: confirm naming before creating, and inform the user they can start working.
