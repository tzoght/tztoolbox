# Review Branch Commits

When the user invokes this command, run the following workflow to review the set of commits on the current feature branch. The goal is to provide a thorough, multi-perspective code review using different AI models, surfacing issues that need human attention.

## 0. Pre-flight checks

Run **all** of these in parallel (they are independent):

1. `git rev-parse --is-inside-work-tree 2>/dev/null` — if not a git repo, stop: **"Not a git repository."**
2. `git branch --show-current` — determine the current branch. If it is `main`, `master`, `develop`, `staging`, `release/*`, or `production`, stop: **"You are on a protected branch. Switch to a feature branch to review its commits."**
3. `git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@'` — detect the default branch. Fall back to `main` if unset. Remember this as `<base-branch>`.
4. `gh --version 2>/dev/null` — remember whether `gh` is available (used for PR context later).

## 1. Determine the review scope

Run **in parallel**:

1. `git log --oneline <base-branch>..HEAD` — list commits on the current branch that are not on the base branch.
2. `git diff --stat <base-branch>...HEAD` — get the diffstat summary.
3. `git diff <base-branch>...HEAD` — get the full diff (this is the primary review input).

If the commit list is empty, stop: **"No commits found on this branch relative to `<base-branch>`. Nothing to review."**

Show the user a summary:

```
📋 Review scope
   Branch:    <current-branch>
   Base:      <base-branch>
   Commits:   <count> commit(s)
   Files:     <file-count> file(s) changed
```

Then list the commits (short hash + message) and the diffstat.

- **If there is exactly one commit** in the list (relative to `<base-branch>`): do **not** ask all vs pick. Tell the user clearly, e.g. **"Only one commit on this branch — reviewing that commit."** Proceed with the full diff already computed (`<base-branch>...HEAD`).
- **If there are two or more commits**: ask the user: **"Review all commits, or select specific ones?"** (all / pick)
  - **all** → proceed with the full diff.
  - **pick** → show the commit list as numbered entries, let the user select one or more. Recompute the diff to include only the selected commits (`git diff <commit>^ <commit>` for each, or a range).

## 2. Multi-model review

Run the following review passes **in parallel** using subagents. Each pass examines the same diff but from a different angle. Use a **fast model** for passes that are pattern-based or structural, and the **default (more capable) model** for passes requiring deeper reasoning.

### Pass 1 — Correctness & Logic (default model)

Prompt the subagent with the full diff and ask it to look for:

- **Logic errors**: off-by-one, nil/null dereference, incorrect conditions, missing edge cases
- **Race conditions or concurrency issues**
- **Error handling gaps**: unchecked errors, swallowed exceptions, missing rollback/cleanup
- **Behavioral regressions**: does the change break existing contracts or documented behavior?
- **Security concerns**: injection, auth bypass, secret leakage, unsafe deserialization

Return findings as a numbered list with file, line range, severity (critical / warning / info), and a short explanation.

### Pass 2 — Code Quality & Style (fast model)

Prompt the subagent with the full diff and ask it to look for:

- **Naming**: unclear, misleading, or inconsistent variable/function/type names
- **Duplication**: repeated logic that should be extracted
- **Complexity**: functions that are too long or deeply nested
- **Dead code**: unreachable branches, unused imports/variables
- **Comment quality**: missing doc comments on public APIs, or stale/misleading comments

Return findings as a numbered list with file, line range, severity (warning / info / nit), and a short explanation.

### Pass 3 — Architecture & Design (default model)

Prompt the subagent with the full diff **and** the list of files changed, and ask it to look for:

- **Separation of concerns**: does the change mix unrelated responsibilities?
- **Dependency direction**: does it introduce circular or inverted dependencies?
- **API surface**: are new public types/functions intentional and well-designed?
- **Testability**: is the new code structured so it can be unit-tested?
- **Backward compatibility**: does it break callers or consumers?

Return findings as a numbered list with file(s), severity (critical / warning / info), and a short explanation.

### Pass 4 — Test Coverage Assessment (fast model)

Prompt the subagent with the full diff and ask it to:

- Identify which **new or changed functions/methods** lack corresponding test changes
- Flag **modified behavior** that has no updated test assertions
- Suggest **specific test cases** that should be added (edge cases, error paths, boundary values)
- Note if any test files were removed or weakened

Return findings as a numbered list with file, what is untested, suggested test description, and severity (warning / info).

## 3. Consolidate findings

After all passes complete, merge results into a single report grouped by severity:

```
🔴 Critical  (<count>)
────────────────────
1. [Correctness] <file>:<lines> — <description>
2. ...

🟡 Warnings  (<count>)
────────────────────
1. [Quality] <file>:<lines> — <description>
2. [Architecture] <file(s)> — <description>
3. ...

🔵 Info / Nits  (<count>)
────────────────────
1. [Quality] <file>:<lines> — <description>
2. [Tests] <file> — <description>
3. ...
```

If there are **no critical findings**, add a note: **"No critical issues found — looking good!"**

At the bottom of the report, show a quick stats line:

```
📊 Review summary: <critical> critical · <warnings> warnings · <info> info/nits — across <pass-count> review passes
```

## 4. Actionable next steps

Based on the findings, suggest concrete next steps:

1. If there are **critical findings**: **"I recommend addressing the critical issues before merging. Want me to help fix any of them?"**
2. If there are **only warnings/info**: **"No blockers found. Consider addressing the warnings when convenient. Ready to proceed?"**
3. If `gh` is available and a PR exists (`gh pr view --json url 2>/dev/null`): **"A PR exists for this branch — want me to post the review summary as a PR comment?"**
4. If `gh` is available and no PR exists: **"Want me to create a PR with this review summary in the description?"**

## 5. Optional: Post review to PR

If the user opts to post the review to a PR:

1. Format the consolidated report as a GitHub-flavored Markdown comment.
2. If a PR exists, run `gh pr comment <pr-number> --body "<review-body>"`.
3. If no PR exists, offer to create one using `gh pr create` with the review summary embedded in the PR body.

Keep the flow conversational: show findings, ask before taking action, and offer to help fix issues.
