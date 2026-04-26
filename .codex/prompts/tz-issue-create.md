# Create Issue

When the user invokes this command, run the following workflow to create a GitHub issue in the current repository.

## 0. Pre-flight checks

Run **all** of these in parallel to minimize latency:

- `git rev-parse --is-inside-work-tree 2>/dev/null` — if not a git repo, stop: **"Not a git repository."**
- `gh auth status` — if not authenticated, stop: **"You are not logged in to GitHub. Run `gh auth login` first."** (This also confirms `gh` is installed; if the binary is missing the command fails.)
- `gh repo view --json nameWithOwner -q .nameWithOwner` — if it fails, stop: **"Could not detect a GitHub repository for this project."** Save the result as `{owner}/{repo}`.

If any check fails, report the **first** blocking error and stop.

## 1. Gather issue details — single prompt

While waiting for the user's response, **prefetch repo metadata in parallel** so it is ready when needed:

- `gh label list --limit 50 --json name -q '.[].name'`
- `gh api repos/{owner}/{repo}/milestones --jq '.[].title'`

Ask the user **one combined question**:

> **Describe the issue.** I need at least a title. You can also include any of the following in your answer:
> - A longer description / body
> - Labels, assignees, or a milestone
>
> Or just give me the title and I'll ask about the rest.

From the user's response:

1. **Title** — extract or suggest a concise title. If the user gave a long description, propose a short title and use the full text as the body.
2. **Body** — use any extra detail the user provided. If none, leave empty.
3. **Labels** — if the user mentioned labels, match them against the prefetched list. If not mentioned, show the available labels as a numbered list and ask: **"Any labels? (numbers, names, or 'none')"**
4. **Assignees** — if the user mentioned assignees, use them (`"me"` maps to `@me`). If not mentioned, default to **none** (do not ask unless the user brought it up).
5. **Milestone** — if milestones exist and the user mentioned one, use it. If milestones exist but weren't mentioned, skip (do not ask). If no milestones exist, skip silently.

The goal: **get to the preview in at most two interactions** (this prompt + one optional follow-up for labels).

## 2. Preview and confirm

Show a compact preview:

```
📌 Title:     <title>
   Labels:    <labels or "none">
   Assignees: <assignees or "none">
   Milestone: <milestone or "none">
   Body:      <first 100 chars or "(empty)">
```

Ask: **"Create this issue?"** (yes / edit / cancel)

- **yes** → proceed to Step 3.
- **edit** → ask what to change, apply it, re-show the preview.
- **cancel** → stop: **"Issue creation cancelled."**

## 3. Create the issue

Run:

```
gh issue create \
  --title "<title>" \
  --body "<body>" \
  [--label "<label>" ...] \
  [--assignee "<user>" ...] \
  [--milestone "<milestone>"]
```

Omit optional flags the user left empty.

## 4. Summary

Show:

- **#<number>** — `<title>`
- **URL**: `<issue-url>`

Then ask: **"Create a branch for this issue?"** If yes, suggest `fix/<number>-<slug>` or `feature/<number>-<slug>` based on labels and offer to run `git checkout -b <branch>`.
