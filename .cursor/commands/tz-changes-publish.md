# Publish Changes

When the user invokes this command, run the following workflow.

## 0. Pre-flight checks

Run all three checks **in parallel** (they are independent):

1. `git status --short` — if the working tree is clean and there is nothing to stage, stop early: **"Nothing to stage — working tree is clean."** Keep this output for Step 1 (do not re-run `git status`).
2. `git branch --show-current` — if the branch is `main`, `master`, or the repo default, **warn the user** and suggest creating a feature branch before continuing. Do not proceed until the user confirms or switches branches.
3. `gh --version 2>/dev/null` — remember whether `gh` is available for Step 3.

## 1. Files to stage

Using the `git status` output already captured in Step 0, present the changed/untracked files as a numbered list.

Ask the user: **"Which files do you want to add to staging?"**

- Accept one or more numbers from the list, specific paths (e.g. `src/foo.ts`, `docs/readme.md`), or **"all"** for everything.
- Run `git add <paths>` or `git add -A` if they said all.
- Show a brief post-add `git status --short` so they can verify what is staged.

## 2. Commit and push

Ask the user: **"Do you want to commit and push?"** (yes / no)

- If **no**: Stop here (staged changes remain; they can commit later). Skip to Step 3 only if `gh` is available.
- If **yes**:
  1. Run `git diff --cached --stat` to get the staged diff summary and generate a suggested commit message (conventional-commit style if the repo follows it, e.g. `feat:`, `fix:`, `docs:`, `chore:`). Present it to the user.
  2. Let the user accept, edit, or replace the suggested message.
  3. Run `git commit -m "<message>"` then `git push` (with `-u origin <branch>` on first push if no upstream is set). Chain these sequentially since push depends on the commit.

## 3. Create a PR

If `gh` is not available (detected in Step 0), inform the user the branch has been pushed, provide the branch name and a suggested PR title, and tell them to open the PR in the browser. Skip the rest of this step.

If `gh` is available, start the existing-PR check **in parallel with the push** from Step 2 (the check queries GitHub, not the local repo):

1. Run `gh pr view --json url 2>/dev/null` to check if a PR already exists for this branch. If one exists, show the URL and ask whether the user wants to open/update it instead of creating a new one.
2. If no existing PR, ask the user: **"Do you want to create a pull request?"** (yes / no)
   - If **no**: Stop here. Inform the user the branch has been pushed.
   - If **yes**: Run `gh pr create --fill` if the user wants defaults, or ask for title/body first and pass them as flags.

Keep the flow conversational: one step at a time, confirm before running destructive or push actions.
