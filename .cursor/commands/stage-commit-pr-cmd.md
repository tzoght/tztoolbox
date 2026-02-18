# Stage, commit & PR

When the user invokes this command, run the following workflow.

## 0. Pre-flight checks

Before starting, perform these checks silently:

- Run `git status`. If the working tree is clean and there is nothing to stage, stop early: **"Nothing to stage — working tree is clean."**
- Run `git branch --show-current`. If the branch is `main`, `master`, or the repo default, **warn the user** and suggest creating a feature branch before continuing (e.g. `git checkout -b feature/<name>`). Do not proceed until the user confirms or switches branches.
- Check whether the GitHub CLI (`gh`) is installed (`gh --version`). Remember the result for Step 3.

## 1. Files to stage

Run `git status` and present the changed/untracked files as a numbered list so the user can see exactly what is available.

Ask the user: **"Which files do you want to add to staging?"**

- Accept one or more numbers from the list, specific paths (e.g. `src/foo.ts`, `docs/readme.md`), or **"all"** for everything.
- Before running `git add`, show a preview summary of what will be staged (new files, modifications, deletions) and ask for confirmation.
- Run `git add <paths>` or `git add -A` if they said all.
- Show the result of `git status` so they can verify.

## 2. Commit and push

Ask the user: **"Do you want to commit and push?"** (yes / no)

- If **no**: Skip to Step 3 (staged changes remain; they can commit later).
- If **yes**:
  1. Generate a suggested commit message by summarizing the staged diff (use conventional-commit style if the repo follows it, e.g. `feat:`, `fix:`, `docs:`, `chore:`). Present it to the user.
  2. Let the user accept, edit, or replace the suggested message.
  3. Run `git commit -m "<message>"` and `git push` (with upstream set if needed, e.g. `git push -u origin <branch>` on first push).

## 3. Create a PR

Ask the user: **"Do you want to create a pull request?"** (yes / no)

- If **no**: Stop here. Inform the user the branch has been pushed and they can open a PR later.
- If **yes**:
  1. Run `gh pr view --json url 2>/dev/null` to check if a PR already exists for this branch. If one exists, show the URL and ask whether the user wants to open/update it instead of creating a new one.
  2. If no existing PR and `gh` is available: run `gh pr create` and follow prompts, or use `gh pr create --fill` if the user wants defaults. If they want to set title/body, ask first or pass flags.
  3. If `gh` is not installed (detected in Step 0): tell the user to open the PR in the browser and provide the branch name and a suggested PR title.

Keep the flow conversational: one step at a time, confirm before running destructive or push actions.
