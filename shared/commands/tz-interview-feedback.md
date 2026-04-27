# Generate Interview Feedback

When the user invokes this command, generate a structured, evidence-based interview feedback report for one or more candidates. The output is a Markdown report (delivered inline in single-candidate mode, written to `feedback.md` per candidate in batch mode). Every claim must be traceable to the provided transcript or resume.

## 0. Pre-flight: detect input mode

This command supports **two modes**. Pick the one that matches what the user provided.

### Mode A — Single candidate (inline / file paths)

The user pasted content, attached files, or supplied individual file paths for one candidate. You need three inputs:

1. **Interview transcript** — full conversation with questions and candidate responses.
2. **Resume / CV** — work history, education, skills, accomplishments.
3. **Interview question list** — the planned questions for this interview.

Ask **once** for anything missing, then proceed with what was provided. Deliver the report inline in the chat.

### Mode B — Batch directory mode

The user gave you a path to a **directory** (e.g., `/path/to/interviews`). Treat it as a batch root and process every immediate sub-directory as one candidate.

**Expected layout** (each candidate gets a folder):

```
<root>/
  <First>_<Last>/
    transcript.txt | transcript.md   # required (prefer .md if both exist)
    questions.txt                    # required
    resume.txt                       # required
    feedback.md                      # written/overwritten by this command
  <Other_Candidate>/
    ...
```

**Per-candidate procedure (Mode B):**

1. Derive the candidate name from the folder name: replace `_` with spaces, preserve casing (e.g., `Ryan_Williamson` → `Ryan Williamson`). If the user supplies a different display name, prefer that.
2. Read the three input files using a tool (Read / fs.readFile / equivalent). If `transcript.md` exists, prefer it; otherwise read `transcript.txt`. `questions.txt` and `resume.txt` are required. File reads are case-sensitive — if the candidate folder uses different casing or extensions (`.markdown`, etc.), use what's there.
3. If any required input is missing or empty, **skip that candidate** and record the reason in the final summary. Do **not** abort the whole batch.
4. Generate the full report per §1–§2 below.
5. **Write the report to `<candidate_dir>/feedback.md`**, overwriting any prior version. The file is the deterministic output of this command — do not prompt before overwriting in Mode B.
6. Each written `feedback.md` must start with the YAML front-matter:

   ```yaml
   ---
   candidate: <derived name>
   source_dir: <absolute path to candidate folder>
   transcript: <transcript filename used>
   generated_by: tz-interview-feedback
   generated_at: <ISO-8601 UTC timestamp>
   ---
   ```

   then a blank line, then the report (starting with the `# Interview Feedback: <Candidate Name>` heading from §2).

**After all candidates are processed**, print a final summary in this exact form (one line per candidate):

```
- <Candidate Name>: OK -> <absolute path to feedback.md>
- <Candidate Name>: SKIPPED (<reason>)
```

Then offer the same follow-ups described in §5.

### Metadata to capture (both modes)

Try to identify (ask only if not obvious from the materials):

- Candidate name — Mode B derives this from the folder name; override if the resume header gives a clearer one.
- Role and seniority (IC, senior, staff, lead, principal, manager, …)
- Interview date and interviewer(s)

If a `~/.cursor/skills/interview-feedback/SKILL.md` or `~/.claude/skills/interview-feedback/SKILL.md` is available, read it for the full framework. Otherwise the procedure below is self-contained.

## 1. Read and map

Work through the materials in this order **before** writing anything:

1. Read the question list first — it defines the bar.
2. Map transcript turns to questions: for every planned question, find where it was (or wasn't) asked and the candidate's answer.
3. Cross-reference the resume:
   - Confirm or contradict concrete claims made in the interview.
   - Flag discrepancies (e.g., resume says "led team of 8", transcript says "supported a team of 8").
   - Note resume accomplishments that were articulated well in the interview.
4. Rate each answer **for the role's level**: `below-bar` / `at-bar` / `above-bar`.
   - **above-bar** — exceeds expectations for the level: unusual depth, originality, scope, or rigor.
   - **at-bar** — meets expectations for the level: complete, accurate, role-appropriate.
   - **below-bar** — falls short for the level: vague, incomplete, off-topic, factually wrong, or shows a gap that matters at this level.

   Be explicit about which level you are calibrating against (IC, senior, staff, lead, principal, manager, …) and apply the same standard to every candidate at that level.
5. Aggregate signals across these dimensions: technical/functional skills, experience & background, communication & soft skills, cultural & team fit, problem-solving & critical thinking.
6. Form the recommendation **last** — strengths and gaps drive it, not overall vibe.

## 2. Generate the report

Produce a single Markdown document with the sections below, in this order.

### Header

Open with a level-1 heading `Interview Feedback: <Candidate Name>` followed by metadata lines:

- **Position:** `<Role>`
- **Interview Date:** `<Date or "Not provided">`
- **Interviewer(s):** `<Names or "Not provided">`

### 2.1 Questions & Answers Summary

For **every** question asked, emit one block in this exact format. Order the blocks in the sequence the questions were asked. Calibrate the rating against the **role's level** — what is `at-bar` for an IC may be `below-bar` for a staff engineer.

```
#### Q<n>: <question text>

**Rating:** below-bar | at-bar | above-bar  _(calibrated for <level>)_
**Answer recap:** <2–4 sentence summary of what the candidate actually said>
**Rationale:** <why the answer earns that rating — cite specific evidence: a short transcript quote, a missing element, or a comparison to the bar for this level>
```

Rating definitions (always relative to the role's level):

- **above-bar** — exceeds expectations for the level: unusual depth, originality, scope, or rigor.
- **at-bar** — meets expectations for the level: complete, accurate, role-appropriate.
- **below-bar** — falls short for the level: vague, incomplete, off-topic, factually wrong, or shows a gap that matters at this level.

After all question blocks, include 1–3 **notable quotes** — short verbatim excerpts that are particularly revealing (positive or negative), each tied to its question number.

### 2.2 Questions Asked by the Candidate

| # | Candidate's question | What it signals |
|---|----------------------|-----------------|
| 1 | … | What this reveals about priorities/thinking |

Add an **Assessment** paragraph: were the questions thoughtful, prepared, role/team focused? Or compensation/perks focused? Did they ask clarifying questions during technical discussions? **If the candidate asked no questions, flag this as a concern.**

### 2.3 Overall Recommendation

Pick exactly one and justify in **2–3 sentences** that name the strongest signal and the most important concern:

- **Strong Hire** — clearly exceeds requirements.
- **Hire** — meets requirements with strong potential.
- **Lean Hire** — meets most requirements; minor concerns.
- **Lean No Hire** — has potential but significant gaps.
- **No Hire** — does not meet key requirements.
- **Strong No Hire** — significantly below requirements or major red flags.

### 2.4 Reasons to Hire

List **3–7** strengths. Each:

```
#### <Strength title>
**Evidence:** <direct quote or resume reference>
**Impact:** <how this benefits the team>
```

### 2.5 Reasons Not to Hire (Gaps)

List every concern, even minor. Each:

```
#### <Gap title>
**Evidence:** <specific observation or missing element>
**Severity:** Critical / Significant / Minor
**Mitigation:** <can this be addressed via training, mentorship, time? how?>
```

### 2.6 Question Coverage

Map planned questions to outcomes:

- ✓ asked and thoroughly addressed
- ◑ asked but partially addressed
- ✗ not asked or not answered

### 2.7 Additional Observations (optional)

Anything material that doesn't fit elsewhere — communication tone, unusual context, suggestions for next-round focus areas.

## 3. Evidence rules (non-negotiable)

- **Every** strength and gap must cite specific evidence: a transcript quote (preferred), a resume line, or an explicit "no evidence found" observation.
- Use **direct quotes** when the candidate's exact words illustrate the point. Keep them short and accurate — never paraphrase inside quotation marks.
- Quantify when possible ("mentioned 3 relevant projects", "5 years of Python per resume").
- If something was **not discussed**, say so under **Question Coverage** or **Additional Observations**. Do **not** infer absence of skill from absence of discussion.

## 4. Do / Don't

**Do:** be objective, specific, balanced, actionable, fair, thorough.

**Don't:** invent evidence, infer undiscussed skills, use vague language, weight non-job-relevant factors, over-index on likeability, soften red flags.

## 5. After delivering the report

Offer concrete follow-ups:

1. "Want me to draft talking points for the debrief meeting?"
2. "Want me to compare this candidate to another you've interviewed?" (Mode B: across the candidates in the batch)
3. "Want me to suggest follow-up questions for the next round?"
4. "Want me to elaborate on any section?"
5. **Mode B only:** "Want me to write a top-level `summary.md` that ranks the candidates side-by-side?"

Keep the flow conversational: deliver the structured report first (Mode A) or the batch summary first (Mode B), then offer to go deeper.
