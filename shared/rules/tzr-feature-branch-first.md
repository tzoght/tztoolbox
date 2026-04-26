# Feature branch for the whole conversation

When planning or starting work, create a feature branch **once at the start** so all edits in this conversation happen on that branch. Follow these steps in order:

## 1. Pre-checks

- Run `git rev-parse --is-inside-work-tree 2>/dev/null`. If this is **not** a git repo, skip all branch logic and proceed normally.
- Run `git branch --show-current`. If the output is empty, the repo is in **detached HEAD** state — warn the user and ask whether to create a branch from the current commit or switch to an existing one.

## 2. Protected branch detection

If the current branch matches any of these patterns, do **not** commit directly to it:

- `main`, `master`, `develop`, `staging`
- `release/*`, `production`

## 3. Create the feature branch

If on a protected branch:

1. Run `git fetch origin` to ensure the local branch is up to date.
2. Check for uncommitted changes with `git status --porcelain`. If dirty, run `git stash` and inform the user their changes have been stashed (they will be re-applied after switching).
3. Derive a branch name from the user's task using a consistent naming convention:
   - `feature/<short-description>` — new functionality
   - `fix/<short-description>` — bug fix
   - `chore/<short-description>` — maintenance, config, tooling
   - `docs/<short-description>` — documentation only
4. Check if the branch already exists locally (`git branch --list <name>`) or remotely (`git ls-remote --heads origin <name>`). If it does, ask the user whether to switch to it or pick a different name.
5. Run `git checkout -b <name>` to create and switch.
6. If changes were stashed in step 2, run `git stash pop` to restore them.

## 4. Stay on the branch

- Do all work on this branch for the rest of the conversation.
- Do **not** create another branch mid-conversation unless the user explicitly asks.

## 5. Already on a feature branch

If the current branch is not a protected branch, stay on it and proceed without creating a new branch.
