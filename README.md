# Code Review Agent 


## Documentation: GitHub Models PR Review Workflow

## Overview

This GitHub Actions workflow is designed to automate **pull request (PR) code reviews** using **GitHub-hosted LLMs** (specifically `gpt-4.1-nano`) via the GitHub Models API. It uses:

* Smart diff file prioritization to manage token limits.
* Markdown-rich prompts for structured review.
* Graceful fallbacks for oversized diffs.

## Objective

To enhance pull request quality checks by automatically generating detailed, structured code reviews that:

* Flag security, performance, and maintainability issues
* Recommend improvements with code snippets
* Provide clear summaries and checklists

---

## Workflow Triggers

| Trigger             | Description                                   |
| ------------------- | --------------------------------------------- |
| `pull_request`      | Runs on PRs targeting the `init-proj` branch. |
| `workflow_dispatch` | Allows manual triggering from GitHub UI.      |

```yaml
ame: GitHub Models PR Review
on:
  pull_request:
    branches:
      - init-proj
  workflow_dispatch:
```

---

## Permissions

This workflow needs to:

* Read the repo content
* Comment on PRs (`pull-requests: write`)
* Create issues if needed
* PAT (Personal Access Token) with Github Models `read-only` permission with `GH_PAT_MODELS` secret for API access

```yaml
permissions:
  contents: read
  pull-requests: write
  issues: write
```

---

## Job: `pr_review`

| Property        | Value              |
| --------------- | ------------------ |
| Runs on         | `ubuntu-latest`    |
| Conditional Run | PR or manual event |

### Step 1: Checkout

Fetches the full repository history:

```yaml
- name: Checkout Repository
  uses: actions/checkout@v4
  with:
    fetch-depth: 0
```

### Step 2: Generate Diff

Creates a diff from base branch to HEAD and saves to `pr_diff.txt`:

```bash
git fetch origin ${{ github.base_ref }}
git diff origin/${{ github.base_ref }}...HEAD > pr_diff.txt
```

---

## Step 3: Smart Diff Prioritization

Handles token limits (OpenAI models have token size limits):

### Logic:

| Condition                 | Action                         |
| ------------------------- | ------------------------------ |
| Diff within token limit   | Use full diff                  |
| Diff exceeds token limit  | Prioritize critical files only |
| Still exceeds safe tokens | Truncate file to \~5000 tokens |

### Categories Used:

| Priority | Category      | File Types                         |
| -------- | ------------- | ---------------------------------- |
| 1        | Core Code     | `.js`, `.py`, `.java`, etc.        |
| 2        | Configuration | `Dockerfile`, `.env`, `.yml`, etc. |
| 3        | Tests/Docs    | `README`, `test_*.py`, etc.        |
| 4        | Styles/Docs   | `.css`, `.md`, `.txt`              |
| 5        | Other         | Anything else                      |

### Snippet: File Categorization

```bash
if echo "$file" | grep -qE '\.(js|jsx|ts|tsx|py|java|...)'; then
  PRIORITY=1
  CATEGORY="Core Code"
```

### Token Calculation (Estimates 1 token ≈ 4 characters):

```bash
ESTIMATED_TOKENS=$((DIFF_SIZE / 4))
```

### Prioritized Files Output:

* Sorted by priority, then file size.
* Only included if total token count stays below `6000`.
* Review summary added to `focused_diff.txt`

### Final Check:

If total tokens still > 6000:

```bash
head -c 20000 pr_diff.txt > truncated.txt
```

---

## Step 4: Review PR with GitHub Models

Uses the GitHub Models API to post a review comment.

### Prompt Structure Sent to Model:

| Section                      | Content Details                       |
| ---------------------------- | ------------------------------------- |
| `## Code Review Summary`     | High-level summary                    |
| `## Critical Issues`         | Blocking problems                     |
| `## Code Quality Analysis`   | Security, Performance, Best Practices |
| `## Detailed Findings`       | Tabular issues                        |
| `## Code Examples`           | Before/After code with explanations   |
| `## Testing Recommendations` | Test ideas                            |
| `## Documentation Notes`     | Documentation suggestions             |

### Model Call Example:

```js
fetch('https://models.github.ai/inference/chat/completions', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${{ secrets.GH_PAT_MODELS }}`,
    ...
  },
  body: JSON.stringify({
    model: "gpt-4.1-nano",
    messages: [...],
    temperature: 0.1,
    max_tokens: 4000
  })
})
```

---

## Fallback: Graceful Failure Handling

If the diff is too large for GPT:

* Parse filenames from diff
* Show summary stats (lines added/removed, file size)
* Post fallback message with:

  * Review checklist
  * Manual review strategy

```markdown
## AI Code Review - Large Diff Analysis
...
- [ ] Check for hardcoded secrets
- [ ] Optimize database queries
...
```

---

## Step 5: Display Final Stats

For debugging and transparency:

```bash
wc -c < pr_diff.txt  # Byte size
wc -l < pr_diff.txt  # Line count
```

Shows if the final diff obeys the token limit (\~6700 tokens max).

---

## Summary Table

| Feature                       | Status       |
| ----------------------------- | ------------ |
| Smart diff prioritization     | ✅ Enabled    |
| File-based categorization     | ✅ Advanced   |
| Token-safe fallbacks          | ✅ Included   |
| Markdown review formatting    | ✅ Structured |
| Graceful API error handling   | ✅ Robust     |
| Final logging and diagnostics | ✅ Verbose    |


---

## Related Files Generated

| File                    | Description                       |
| ----------------------- | --------------------------------- |
| `pr_diff.txt`           | Full or focused diff for LLM      |
| `focused_diff.txt`      | Diff after smart prioritization   |
| `sorted_files.txt`      | File list sorted by priority/size |
| `priority_analysis.txt` | Token and category for each file  |

## Setup

1. **Create a Fine-Grained Personal Access Token:**
   - Go to GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
   - Grant these permissions:
     - **Repository:** Contents (read), Issues (write), Pull requests (write), Metadata (read)
     - **Account:** GitHub Models (read)

2. **Add to Repository Secrets:**
   ```
   GH_FINE_GRAINED_PAT = your_token_here
   ```

3. **Use in Workflow:**
   ```yaml
   - uses: your-username/code-review-agent@v1
     with:
       github-token: ${{ secrets.GH_FINE_GRAINED_PAT }}
   ```

## Installation

Add this to your workflow file:

```yaml
- uses: your-username/code-review-agent@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    model: 'gpt-4.1-nano'
```
