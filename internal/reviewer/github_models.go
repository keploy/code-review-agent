package reviewer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GitHubModelsClient struct {
    token      string
    httpClient *http.Client
}

type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Messages    []ChatMessage `json:"messages"`
    Model       string        `json:"model"`
    Temperature float64       `json:"temperature"`
    MaxTokens   int           `json:"max_tokens"`
}

type ChatResponse struct {
    Choices []struct {
        Message ChatMessage `json:"message"`
    } `json:"choices"`
}

func NewGitHubModelsClient(token string) *GitHubModelsClient {
    return &GitHubModelsClient{
        token:      token,
        httpClient: &http.Client{},
    }
}

func (c *GitHubModelsClient) GenerateReview(diff, model string, temperature float64, maxTokens int) (string, error) {
    prompt := c.buildPrompt(diff)
    
    request := ChatRequest{
        Messages: []ChatMessage{
            {
                Role:    "system",
                Content: "You are a senior software engineer conducting thorough code reviews. Provide detailed, actionable feedback with proper markdown formatting, concrete code examples, and structured recommendations. Focus on security, performance, maintainability, and best practices. Always include before/after code examples for suggested improvements.",
            },
            {
                Role:    "user",
                Content: prompt,
            },
        },
        Model:       model,
        Temperature: temperature,
        MaxTokens:   maxTokens,
    }

    jsonData, err := json.Marshal(request)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %w", err)
    }

    req, err := http.NewRequest("POST", "https://models.github.ai/inference/chat/completions", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "GitHub-Actions-PR-Review/1.0")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        // MISSING: Enhanced fallback like YAML
        if resp.StatusCode == 413 || strings.Contains(string(body), "tokens_limit_reached") {
            return c.generateLargeDiffFallback(diff), nil
        }
        return "", fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
    }

    var response ChatResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return "", fmt.Errorf("failed to unmarshal response: %w", err)
    }

    if len(response.Choices) == 0 {
        return "", fmt.Errorf("no choices in response")
    }

    return response.Choices[0].Message.Content, nil
}

func (c *GitHubModelsClient) buildPrompt(diff string) string {
    return fmt.Sprintf(`You are an expert code reviewer. Analyze the provided git diff and deliver a comprehensive, professional code review following the exact structure below.

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
` + "```diff\n" + diff + "\n```")
}

func (c *GitHubModelsClient) generateLargeDiffFallback(diff string) string {
    // Parse changed files from diff
    lines := strings.Split(diff, "\n")
    var changedFiles []string
    
    for _, line := range lines {
        if strings.HasPrefix(line, "diff --git") {
            // Extract filename: "diff --git a/file.go b/file.go"
            parts := strings.Fields(line)
            if len(parts) >= 4 {
                file := strings.TrimPrefix(parts[3], "b/")
                changedFiles = append(changedFiles, file)
            }
        }
    }
    
    addedLines := 0
    removedLines := 0
    for _, line := range lines {
        if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
            addedLines++
        }
        if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
            removedLines++
        }
    }
    
    return fmt.Sprintf(`## AI Code Review - Large Diff Analysis

This pull request contains extensive changes that exceed optimal token limits for detailed AI analysis.

### Change Overview
- **Files Modified:** %d
- **Lines Added:** ~%d
- **Lines Removed:** ~%d
- **Diff Size:** %dKB (~%d tokens)

### Files Requiring Manual Review

#### High Priority Files
%s

### Recommended Manual Review Checklist

#### 🔒 Security Review
- [ ] Check for hardcoded secrets, API keys, or passwords
- [ ] Validate input sanitization and XSS prevention  
- [ ] Review authentication and authorization changes

#### ⚡ Performance Review  
- [ ] Look for inefficient database queries or N+1 problems
- [ ] Check for memory leaks in loops or event handlers
- [ ] Review caching strategies and implementation

#### 🧪 Testing Requirements
- [ ] Ensure unit tests cover new functionality
- [ ] Add integration tests for API changes
- [ ] Update end-to-end tests for UI modifications

---
**Review Status:** Manual review required due to diff complexity  
**Fallback Reason:** Token limit exceeded`, 
        len(changedFiles), addedLines, removedLines, 
        len(diff)/1024, len(diff)/4,
        c.formatFileList(changedFiles))
}

func (c *GitHubModelsClient) formatFileList(files []string) string {
    var result strings.Builder
    for _, file := range files {
        if strings.Contains(file, ".go") || strings.Contains(file, ".js") || 
           strings.Contains(file, ".py") || strings.Contains(file, ".java") {
            result.WriteString(fmt.Sprintf("- `%s` - Core application logic\n", file))
        }
    }
    return result.String()
}