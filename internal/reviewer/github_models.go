package reviewer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
    "github.com/keploy/code-review-agent/internal/config"
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
    prompt := BuildReviewPrompt(diff)
    
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
        if resp.StatusCode == 413 || strings.Contains(string(body), "tokens_limit_reached") {
            return "", config.ErrTokenLimitExceeded
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
