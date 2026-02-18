# Delete Feature Branch

When the user invokes this command, run the following workflow to safely delete a feature branch both locally and remotely.

## 0. Pre-flight checks

- Run `git rev-parse --is-inside-work-tree 2>/dev/null`. If not a git repo, stop: **"Not a git repository."**
- Run `git branch --show-current` to determine the current branch.
- Run `git fetch --prune` to sync with the remote and clean up stale tracking refs.

## 1. Select the branch to delete

Ask the user: **"Which branch do you want to delete?"**

- List local branches with `git branch` (exclude `main`, `master`, `develop`, `staging`, and `release/*`).
- Present them as a numbered list so the user can pick by number or type a name.
- If the user provides a branch name directly, use it.
- **Refuse to delete protected branches** (`main`, `master`, `develop`, `staging`, `release/*`, `production`). If the user picks one, warn them and ask again.

## 2. Safety checks

Before deleting, perform the following:

1. **Merged status**: Run `git branch --merged main` (or the default branch) to check if the selected branch has been fully merged. If **not merged**, warn the user: **"This branch has not been merged. Deleting it will lose unmerged commits."** Ask for explicit confirmation to proceed.
2. **Open PR check**: If `gh` is available, run `gh pr list --head <branch> --json number,url,state` to check for open PRs. If an open PR exists, warn the user and show the PR URL. Ask whether to close the PR and continue, or abort.
3. **Current branch conflict**: If the user is currently on the branch they want to delete, switch to the default branch first (`git checkout main` or `master`).

## 3. Delete the branch

After the user confirms:

1. **Delete locally**: Run `git branch -d <branch>` (safe delete). If the branch is not fully merged and the user confirmed in Step 2, use `git branch -D <branch>` (force delete).
2. **Delete remotely**: Run `git push origin --delete <branch>`. If the remote branch doesn't exist, skip silently.

## 4. Summary

Show a recap:

- **Deleted locally**: yes / no (with reason if skipped)
- **Deleted remotely**: yes / no (with reason if skipped)
- **Current branch**: show which branch the user is now on

Keep the flow conversational: confirm before every destructive action.
