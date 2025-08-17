package reviewer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type OllamaClient struct {
    baseURL    string
    httpClient *http.Client
    model      string
}

type OllamaRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
}

type OllamaResponse struct {
    Response string `json:"response"`
    Done     bool   `json:"done"`
}

func NewOllamaClient(model string) *OllamaClient {
    return &OllamaClient{
        baseURL: "http://localhost:11434",
        httpClient: &http.Client{
            Timeout: 4000 * time.Second, // 4000 seconds for large models
        },
        model: model,
    }
}

func (c *OllamaClient) SetupOllama() error {
    fmt.Println("🔧 Setting up Ollama...")

    // Install Ollama
    cmd := exec.Command("bash", "-c", "curl -fsSL https://ollama.com/install.sh | sh")
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to install Ollama: %w (output: %s)", err, string(output))
    }

    // --- IMPROVEMENT: Verify installation and use absolute path ---
    ollamaPath := "/usr/local/bin/ollama"
    if _, err := os.Stat(ollamaPath); os.IsNotExist(err) {
        return fmt.Errorf("ollama binary not found at %s after installation. The install script might have failed or installed it elsewhere", ollamaPath)
    }
    fmt.Printf("✅ Ollama binary found at %s\n", ollamaPath)
    // --- END IMPROVEMENT ---

    fmt.Println("🚀 Starting Ollama service with logging...")
    logFile, err := os.Create("ollama_serve.log")
    if err != nil {
        return fmt.Errorf("failed to create ollama log file: %w", err)
    }
    defer logFile.Close()

    // Start the service using the absolute path
    cmd = exec.Command(ollamaPath, "serve")
    cmd.Stdout = logFile
    cmd.Stderr = logFile

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start Ollama service: %w", err)
    }

    // Wait for service to be ready
    fmt.Println("⏳ Waiting for Ollama service to be ready...")
    serviceReady := false
    for i := 0; i < 60; i++ {
        if c.isServiceReady() {
            serviceReady = true
            break
        }
        time.Sleep(2 * time.Second)
    }

    if !serviceReady {
        logContent, _ := os.ReadFile("ollama_serve.log")
        errorMsg := fmt.Sprintf("Ollama service failed to start within timeout. Log output:\n---\n%s\n---", string(logContent))
        return fmt.Errorf(errorMsg)
    }

    // Pull model using the absolute path
    fmt.Printf("📥 Pulling model %s...\n", c.model)
    cmd = exec.Command(ollamaPath, "pull", c.model)
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to pull model %s: %w (output: %s)", c.model, err, string(output))
    }

    fmt.Println("✅ Ollama setup complete")
    return nil
}

func (c *OllamaClient) isServiceReady() bool {
    resp, err := c.httpClient.Get(c.baseURL + "/api/tags")
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}

func (c *OllamaClient) GenerateReview(diff string) (string, error) {
    prompt := c.buildPrompt(diff)
    
    request := OllamaRequest{
        Model:  c.model,
        Prompt: prompt,
        Stream: false,
    }
    
    jsonData, err := json.Marshal(request)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %w", err)
    }
    
    req, err := http.NewRequest("POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("Ollama API error: %d - %s", resp.StatusCode, string(body))
    }
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response: %w", err)
    }
    
    var response OllamaResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return "", fmt.Errorf("failed to unmarshal response: %w", err)
    }
    
    return c.cleanResponse(response.Response), nil
}

func (c *OllamaClient) buildPrompt(diff string) string {
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
**Include a small Mermaid sequence diagram summarizing the PR.** 

Example Mermaid Diagram:
`+"```mermaid\n"+`sequenceDiagram
    Mermaid code the diagram should show the concise logical changes in the codebase between old code and new code.
`+"```\n"+`

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


---

Here is the diff to review:
` + "```diff\n" + diff + "\n```")
}

func (c *OllamaClient) cleanResponse(response string) string {
    // Remove thinking process if present
    lines := []string{}
    skipThinking := false
    
    for _, line := range []string{response} {
        if line == "Thinking..." {
            skipThinking = true
            continue
        }
        if line == "...done thinking." {
            skipThinking = false
            continue
        }
        if !skipThinking {
            lines = append(lines, line)
        }
    }
    
    if len(lines) > 0 {
        return lines[0]
    }
    return response
}

func (c *OllamaClient) Cleanup() {
    fmt.Println("🧹 Cleaning up Ollama service...")
    exec.Command("pkill", "-f", "ollama serve").Run()
}