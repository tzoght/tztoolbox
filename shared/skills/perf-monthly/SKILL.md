---
name: perf-monthly
description: >-
  Runs a monthly performance tracker update for an EM's IC direct reports using
  Jira, GitHub, and qualitative context. Use for performance tracker drafts.
---

# Monthly Performance Tracker Update

Run the monthly performance tracker update for an EM's IC direct reports.

This skill writes to a shared tracker, typically in Notion. It updates existing rows only and must not create new pages. Always require explicit confirmation before writing to the tracker.

The workflow pulls signal from two required sources, Jira activity and GitHub PR activity over the last 30 days, plus any qualitative context the EM keeps about ICs, such as people files, 1:1 summaries, or team context files. It produces evidence-based drafts and interviews the EM to determine ratings. It never suggests Performance Activity or Trend itself.

## Setup

Before the first run, ask the user to provide or confirm this configuration. Keep the configuration in the conversation or in a local user-owned context file; do not commit private organization details.

Required settings:

| Setting | Value | Notes |
|---------|-------|-------|
| `IC_SCOPE` | `[Name 1, Name 2, ...]` | Full names of the EM's direct-report ICs. |
| `EMAIL_DOMAIN` | `[company-domain.example]` | Used to derive work emails for Jira lookup, for example `firstname.lastname@[EMAIL_DOMAIN]`. |
| `GITHUB_ORG` | `[github-org]` | GitHub organization for PR and review queries. |
| Tracker database | `[notion-tracker-database-id-or-url]` | Shared performance tracker database. |

Optional qualitative context sources:

| Setting | Value | Notes |
|---------|-------|-------|
| `PEOPLE_FILE_PATH` | `people/[ic-name].md` | Local markdown files per IC, if maintained. Leave blank if unused. |
| 1:1 database | `[notion-1on1-database-id-or-url]` | Notion database of 1:1 summaries. Adapt the sample query to the user's schema. |
| `CONTEXT_FILES` | `[paths to context files]` | Rubrics, role expectations, team notes, or other context files. Leave blank if unused. |
| GitHub username convention | `[describe convention]` | Optional naming convention for deriving usernames. If absent, search by name. |
| Other qualitative sources | `[describe source and loading steps]` | Obsidian vaults, exported notes, Slack saves, shared docs, or other sources. |

If the user does not keep structured qualitative context, skip Phase 1a. The skill still works from Jira and GitHub signal, and Phase 3 collects supplementary context from the EM.

Jira requires no project key by default. Query across the connected Jira instance by assignee unless the user provides a narrower scope.

## Instructions

### Phase 1: Load Context

#### 1a. Qualitative context in parallel

Load whatever qualitative context the user keeps for each IC. This phase is optional and depends on the user's note-taking taxonomy. Configure the sub-steps below for the sources they actually use and skip the rest.

Run all configured sources in parallel.

Source: local context and people files, if `CONTEXT_FILES` or `PEOPLE_FILE_PATH` are set:

- Read all `CONTEXT_FILES`.
- Read one `PEOPLE_FILE_PATH` per IC in `IC_SCOPE`.

People files often contain curated context, such as recent review notes, strengths, growth areas, open items, or abridged 1:1 logs. Treat them as useful synthesis, not the canonical 1:1 record.

Source: Notion 1:1 summaries database, if configured:

Query the database for the past 30 days for each in-scope IC. Adapt this sample query to the database's actual columns and naming:

```sql
SELECT url, "Name", "date:date:start" AS date, "topics", "relationship"
FROM "[1:1 database]"
WHERE "date:date:start" >= '[DATE_30_DAYS_AGO]'
  AND "Name" IN ('<shorthand 1>', '<shorthand 2>', ...)
  AND "relationship" = 'Direct Report'
ORDER BY "date:date:start" DESC
```

Schema assumptions to verify or replace:

- A title-style column matching whatever shorthand the EM uses for each IC in 1:1 entries.
- A date column for filtering to the past 30 days.
- Optionally, a relationship or category column to filter direct reports.
- Optionally, a topics or tags column to flag performance-management entries.

Fetch each result URL to read the actual 1:1 content. Treat raw 1:1 summaries as canonical. Prefer them over curated people-file 1:1 logs when they conflict.

Source: anything else the user keeps:

If the user tracks 1:1 notes or IC context in another system, load it according to the source they provide. The skill is source-agnostic; any relevant context should be folded into the synthesis in Phase 4.

Across all sources, skip queries for any IC on extended leave if the configuration or people file flags them as on leave.

#### 1b. External activity lookups in parallel across ICs

Run this lookup step in parallel with Phase 1a. For each in-scope IC, skipping anyone on leave:

1. Derive the work email as `firstname.lastname@[EMAIL_DOMAIN]`, unless the user provides an explicit email mapping.
2. Resolve the Jira account with the Jira MCP account lookup tool using the work email.
3. Resolve the GitHub username using the configured convention, if any, then verify it with `gh api "users/<username>" --jq '{login, name}'`.
4. If no convention is configured or verification fails, search by name in the org:

```bash
gh api "search/users?q=<full+name>+org:[GITHUB_ORG]" --jq '.items[] | {login, name}'
```

If neither GitHub strategy returns a match, note the gap and proceed without GitHub data for that IC.

After accounts are resolved, run these activity queries for each IC in parallel:

Jira tickets closed this month:

```text
assignee = "[accountId]" AND status changed to Done AFTER -30d ORDER BY updated DESC
```

Jira active tickets:

```text
assignee = "[accountId]" AND status not in (Done, Closed, Resolved) AND updated >= -30d ORDER BY updated DESC
```

Some Jira workflows use completion statuses such as `Merged` or `Released` that map to a done status category but are not literally named `Done`. When synthesizing in Phase 4, classify by `statusCategory.key` (`done`, `indeterminate`, `new`) rather than by literal status name.

GitHub PRs authored:

```bash
gh search prs --owner=[GITHUB_ORG] --author=<username> --created=[DATE_30_DAYS_AGO]..[TODAY] --json number,title,repository,state,createdAt --limit 50
```

GitHub PR reviews given:

```bash
gh search prs --owner=[GITHUB_ORG] --reviewed-by=<username> --updated=[DATE_30_DAYS_AGO]..[TODAY] --json number,title,repository,state --limit 50
```

Retain raw counts and notable items, such as overdue Jira tickets, PRs open for more than two weeks, or unusually high or low review volume relative to the IC's pattern.

### Phase 2: Determine the Current Month Column

Query the tracker schema to confirm the current month's notes column name:

```sql
SELECT * FROM "[Tracker database]" LIMIT 1
```

Many performance trackers use a per-month notes column following a pattern like `[status marker] YYYY-MM (Mon) Notes`, for example `2026-04 (Apr) Notes`. Use today's date to identify the correct column. If the column for the current month does not exist, stop and ask the EM how to proceed before continuing.

### Phase 3: Per-IC Supplementary Input

Before drafting, ask the EM for any additional context per IC that may not be in the configured sources, such as Slack conversations, ad-hoc feedback, delivery signals, or stakeholder input from the past month.

Ask one IC at a time, in `IC_SCOPE` order:

> "Anything to add for [Name] this month? Paste Slack threads, notes, URLs, or anything else, or press enter to skip."

Accept free-form input. Fold whatever is provided into that person's draft alongside people-file context and external signals. Do not require structured input.

### Phase 4: Synthesize Evidence

For each IC, synthesize from their people file, recent 1:1 summaries, Jira and GitHub activity, and supplementary input. Produce an evidence summary and draft these fields:

- **Areas of Coaching / Concern / Needs Improvement**: 1-2 bullets. For stronger performers, these may be growth edges rather than concerns.
- **Next Steps**: the active coaching or development focus for the coming month.
- **Monthly Notes**: 2-3 sentences summarizing the month's signal, including what was delivered, notable behavior, and where things stand.

Do not suggest or imply a Performance Activity rating or Trend. Those are the manager's call and are determined in Phase 5.

Evidence synthesis guidance:

- Use Jira and GitHub data as a delivery baseline alongside qualitative signals.
- Prefer 1:1 summaries and people-file context first, external signals second, and supplementary input third.
- A high closed-ticket count with no 1:1 signal means "delivering but uncalibrated"; surface it as a data point, not a conclusion.
- An open PR sitting for more than two weeks is worth surfacing in Next Steps or Monthly Notes if it appears stuck.
- Low GitHub review volume relative to the IC's pattern is a soft signal, not a hard concern.
- Jira active-ticket staleness, such as tickets not updated in more than two weeks, is a flag worth surfacing, not an automatic downgrade.
- If Jira or GitHub data could not be retrieved for an IC, note the gap explicitly.

Keep entries short. Review existing tracker entries for calibration on length and tone. These are executive-readable summaries, not detailed reviews.

### Phase 5: Manager Interview for Performance Activity and Trend

For each IC, one at a time:

1. Present the evidence summary and drafted fields.
2. Ask: "What's your Performance Activity rating for [Name]? (Growth / Performing / Watching / Needs Improvement)"
3. Ask: "What's the Trend? (Up / Flat / Down)"
4. If the answer conflicts with the evidence, push back directly before accepting it. Name the specific tension, for example:

   > "You said Growth, but Jira shows only three tickets closed this month and the recent 1:1 notes mention delivery slowdowns. What are you seeing that makes this Growth rather than Performing?"

   > "You said Trend is Up, but there is no 1:1 signal this month and GitHub review volume dropped meaningfully. What's driving the Up read?"

   Accept an answer that conflicts with evidence only when the manager names something specific that the data does not capture. Do not accept vague reassurance as sufficient.
5. Once both values are confirmed for an IC, move to the next.

After all ICs are done, display the complete table:

```markdown
| Name | Performance Activity | Trend | Areas of Coaching | Next Steps | Monthly Notes |
|------|----------------------|-------|-------------------|------------|---------------|
```

Then ask:

> "Does this look right? Any corrections before I push to the tracker?"

Do not write to the tracker until explicitly confirmed. This may be a shared database visible to leadership.

### Phase 6: Write to the Tracker

#### 6a. Offer a local markdown draft first

Before writing to the tracker, ask:

> "Want me to write the confirmed entries to a local markdown file first so you can review and edit before it goes to the tracker?"

If yes, write to `outputs/YYYY-MM-DD-perf-tracker-[Mon]-draft.md` with all fields per IC:

```markdown
# Perf Tracker Draft - [Month] [Year]

## [IC Name]

**Performance Activity:** [value]
**Trend:** [value]
**Areas of Coaching / Concern / Needs Improvement:**
[bullets]
**Next Steps:**
[text]
**Monthly Notes:**
[text]

---
```

Then say:

> "File written to [path]. Make any edits directly in the file, then let me know when you're ready and I'll read it back before pushing to the tracker."

Read the file back in before proceeding. Use the file contents as the final source of truth. If anything changed from the confirmed draft, note the delta and ask whether the changes are intentional before writing.

#### 6b. Write to Notion or the configured tracker

For each IC:

1. Find the existing tracker row by querying the tracker database for the IC's name.
2. Update only the existing row with:
   - `Performance Activity`
   - `Areas of Coaching / Concern / Needs Improvement`
   - `Trend`
   - `Next Steps`
   - Current month's notes column

If a row is not found for an IC, report it rather than creating a new row. Row creation is out of scope for this skill.

Report each successful write. If any write fails, report the error and the field values so the EM can update manually.

## Notes

- Skip ICs on extended leave, such as parental, medical, or sabbatical leave. Flag leave status in the people file or configuration.
- This skill is for IC direct reports only. Do not include EMs, Team Leads, or indirect reports.
- If people files are stale, for example no 1:1 log entries in more than 30 days, note the gap in the draft and flag it to the EM.
- If the user keeps one document for all ICs instead of one people file per IC, adapt Phase 1a accordingly.
