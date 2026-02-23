# Delete Branch

When the user invokes this command, run the following workflow to safely delete a feature branch both locally and remotely.

## 0. Pre-flight checks

Run all four commands **in parallel** (they are independent). Having the branch list ready up front avoids an extra round-trip before prompting the user.

1. `git rev-parse --is-inside-work-tree 2>/dev/null` — if not a git repo, stop: **"Not a git repository."**
2. `git branch --show-current` — determine the current branch.
3. `git fetch --prune origin` — sync remote-tracking refs and prune stale ones.
4. `git branch` — list local branches (used in Step 1).

## 1. Select the branch to delete

Ask the user: **"Which branch do you want to delete?"**

- From the branch list already retrieved in Step 0, exclude `main`, `master`, `develop`, `staging`, `release/*`, and `production`.
- Present the remaining branches as a numbered list so the user can pick by number or type a name.
- If the user provides a branch name directly, use it.
- **Refuse to delete protected branches.** If the user picks one, warn them and ask again.

## 2. Safety checks

Run **all** of the following checks **in parallel** (they are independent reads). If `gh` is available, combine both PR queries into a single call.

1. **Merged status** (if **any** method indicates merged, treat the branch as merged):
   - **Local git check**: `git branch --merged main` (or the default branch) — is the branch tip reachable from `main`?
   - **Remote git check**: `git branch -r --merged origin/main` — is the remote-tracking branch merged? (covers cases where local `main` is behind).
   - **GitHub PR check**: If `gh` is available, run `gh pr list --head <branch> --json number,url,state` (returns all states) and filter for `merged` entries. This single call replaces two separate `gh pr list` invocations and covers squash/rebase merges where commit SHAs differ.
   - If **none** indicate merged, warn the user: **"This branch has not been merged. Deleting it will lose unmerged commits."** Ask for explicit confirmation.
   - If **any** confirms merged, inform the user: **"This branch has been merged (via PR or directly)."** and proceed normally.
2. **Open PR check**: From the same `gh pr list` result above, filter for `open` entries. If an open PR exists, warn the user and show the PR URL. Ask whether to close the PR and continue, or abort.
3. **Current branch conflict**: If the user is currently on the branch they want to delete (known from Step 0), switch to the default branch first (`git checkout main` or `master`).

## 3. Delete the branch

After the user confirms:

1. **Delete locally**: Run `git branch -d <branch>` (safe delete). If the branch is not fully merged and the user confirmed in Step 2, use `git branch -D <branch>` (force delete).
2. **Delete remotely**: Reuse the merged-status result from Step 2 (do not re-check).
   - If the remote branch **is merged**: run `git push origin --delete <branch>`.
   - If the remote branch **is not merged**: warn the user: **"The remote branch has not been merged into origin/main. Deleting it remotely will lose any pushed commits that haven't been merged."** Ask for explicit confirmation before proceeding.
   - If the remote branch doesn't exist (check with `git branch -r --list origin/<branch>` from the already-fetched cache), skip silently.

## 4. Summary

Show a recap:

- **Deleted locally**: yes / no (with reason if skipped)
- **Deleted remotely**: yes / no (with reason if skipped)
- **Current branch**: show which branch the user is now on

Keep the flow conversational: confirm before every destructive action.
