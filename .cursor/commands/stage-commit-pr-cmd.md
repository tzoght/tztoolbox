# Stage, commit & PR

When the user invokes this command, run the following workflow.

## 1. Files to stage

Ask the user: **"Which files do you want to add to staging?"**

- Accept a list of paths (e.g. `src/foo.ts`, `docs/readme.md`), or **"all"** for everything.
- Run `git add <paths>` or `git add -A` if they said all. If they gave paths, use those exactly.
- Show the result of `git status` so they can confirm.

## 2. Commit and push

Ask the user: **"Do you want to commit and push?"** (yes / no)

- If **yes**: Ask for a commit message, then run `git commit -m "<message>"` and `git push` (with upstream set if needed, e.g. `git push -u origin <branch>` on first push).
- If **no**: Skip to step 3 (staged changes remain; they can commit later).

## 3. Create a PR

Create a pull request from the current branch:

- If the GitHub CLI (`gh`) is available: run `gh pr create` and follow prompts (or use `gh pr create --fill` if they want to use defaults). If they want to set title/body, ask first or pass flags.
- If `gh` is not installed: tell the user to open the PR in the browser (e.g. from the repo’s "Compare & pull request" link) and optionally give the branch name and a one-line suggestion for the PR title.

Keep the flow conversational: one step at a time, confirm before running destructive or push actions.
