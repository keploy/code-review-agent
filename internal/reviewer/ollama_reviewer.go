package reviewer

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
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
            Timeout: 300 * time.Second, // 5 minutes for large models
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
    
    // Start Ollama service
    fmt.Println("🚀 Starting Ollama service...")
    cmd = exec.Command("nohup", "ollama", "serve")
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start Ollama service: %w", err)
    }
    
    // Wait for service to be ready
    fmt.Println("⏳ Waiting for Ollama service to be ready...")
    for i := 0; i < 30; i++ {
        if c.isServiceReady() {
            break
        }
        time.Sleep(2 * time.Second)
    }
    
    if !c.isServiceReady() {
        return fmt.Errorf("Ollama service failed to start within timeout")
    }
    
    // Pull model
    fmt.Printf("📥 Pulling model %s...\n", c.model)
    cmd = exec.Command("ollama", "pull", c.model)
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
    return fmt.Sprintf(`You are an expert code reviewer. Your task is to analyze the provided git diff, identify potential issues, suggest improvements, and summarize the key changes. Pay close attention to:
- Readability and code style adherence
- Potential bugs or edge cases  
- Security vulnerabilities
- Performance implications
- Adherence to best practices
- Missing tests or documentation

Here is the git diff:
` + "```diff\n" + diff + "\n```" + `

Provide your review in a concise and actionable manner, using markdown formatting including code blocks where necessary. Start with a brief summary, then list specific findings and suggestions. Output only the review comments for the code changes. Do not include your reasoning or thinking process.
End with a summary of the most critical issues found.`)
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