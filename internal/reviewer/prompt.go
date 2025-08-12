package reviewer

func BuildReviewPrompt(diff string) string {
    return `You are an expert code reviewer. Analyze the provided git diff and deliver a comprehensive, professional code review following the exact structure below.

### FORMATTING REQUIREMENTS:
- Use proper markdown with clear sections
- Include specific code snippets with language tags
- Provide concrete examples for improvements
- Use tables for structured findings
- Reference specific file locations
- Professional tone, no emojis

### REQUIRED STRUCTURE:

## Code Review Summary
Brief overview of changes and overall quality assessment.

## Critical Issues
List high-priority issues requiring immediate attention with clear impact explanations.

## Code Quality Analysis

### Security Concerns
Identify security issues with code examples and explanations.

### Performance Issues  
Highlight performance problems with optimization suggestions.

### Best Practices
Note coding standard violations and improvement opportunities.

## Detailed Findings

<details>
  <summary>📂 Click to expand issue table</summary>

| Category | Issue Description | Location (File:Line) | Severity | Recommendation |
|----------|-------------------|----------------------|----------|----------------|
| Example  | Description here  | file.js:42           | High     | Specific fix   |

</details>

## Code Examples

### Current Implementation
Show problematic code snippets with explanations of why they're issues.

### Suggested Improvements
Present corrected versions with detailed explanations.

## Testing Recommendations
Specific test suggestions for the changes.

## Documentation Notes
Documentation improvements or additions needed.

---

Here is the diff to review:
` + "```diff\n" + diff + "\n```"
}