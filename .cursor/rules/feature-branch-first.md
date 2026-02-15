---
description: Create a feature branch before modifying code
alwaysApply: true
---

# Feature branch before edits

Before making any code modifications, create a feature branch:

1. Run `git branch --show-current`. If the branch is `main`, `master`, or the repo default, create and switch to a new branch (e.g. `feature/short-description`, `fix/issue-name`) with `git checkout -b <name>`.
2. Then proceed with the requested code changes.
3. If the user is already on a feature branch, proceed without creating a new branch.