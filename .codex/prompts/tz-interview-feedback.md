# Generate Interview Feedback

When the user invokes this command, generate a structured, evidence-based interview feedback report for a candidate. The output must be a single Markdown report in the format defined below. Every claim must be traceable to the provided transcript or resume.

## 0. Pre-flight: collect inputs

You need three inputs. Ask **once** for anything missing, then proceed with whatever was provided.

1. **Interview transcript** — full conversation with questions and candidate responses (paste, attached file, or path).
2. **Resume / CV** — work history, education, skills, accomplishments.
3. **Interview question list** — the planned questions for this interview.

Also try to capture (ask only if not obvious from the materials):

- Candidate name
- Role and seniority (IC, senior, staff, lead, manager, etc.)
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
4. Score each answer: **Strong** / **Adequate** / **Weak**.
   - **Strong** — comprehensive, specific, showed depth and relevant experience.
   - **Adequate** — answered but lacked depth or specificity.
   - **Weak** — vague, off-topic, or failed to address the question.
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

A table with every question asked and how the candidate responded:

| # | Question | Answer summary | Quality |
|---|----------|----------------|---------|
| 1 | … | Key points from the response | Strong / Adequate / Weak |

Below the table, include 1–3 **notable quotes** — short verbatim excerpts that are particularly revealing (positive or negative), each tied to its question number.

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
2. "Want me to compare this candidate to another you've interviewed?"
3. "Want me to suggest follow-up questions for the next round?"
4. "Want me to elaborate on any section?"

Keep the flow conversational: deliver the structured report first, then offer to go deeper.
