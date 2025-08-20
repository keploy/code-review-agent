package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Config struct {
    GitHubToken     string
    ModelsToken     string
    Model           string
    MaxTokens       int
    Temperature     float64
    IgnorePatterns  []string
    IncludePatterns []string
    RepoOwner       string
    RepoName        string
    PRNumber        int
    BaseRef         string
    HeadRef         string

    // Ollama settings
    OllamaModel    string
    EnableOllama   bool
    UseOllamaFallback bool
}

func LoadFromEnv() (*Config, error) {
    cfg := &Config{
        // Try both hyphen and underscore versions
       GitHubToken:      getEnvValue("", "INPUT_GITHUB-TOKEN", "INPUT_GITHUB_TOKEN").(string),
        ModelsToken:     getEnvValue("", "INPUT_GITHUB-TOKEN", "INPUT_GITHUB_TOKEN").(string),
        Model:           getEnvValue("gpt-4o-mini", "INPUT_MODEL").(string),
        MaxTokens:       getEnvValue(8000, "INPUT_MAX-TOKENS", "INPUT_MAX_TOKENS").(int),
        Temperature:     getEnvValue(0.2, "INPUT_TEMPERATURE", "INPUT_TEMPERATURE").(float64),
        IgnorePatterns:  strings.Split(getEnvValue("", "INPUT_IGNORE-PATTERNS", "INPUT_IGNORE_PATTERNS").(string), ","),
        IncludePatterns: strings.Split(getEnvValue("", "INPUT_INCLUDE-PATTERNS", "INPUT_INCLUDE_PATTERNS").(string), ","),

        RepoOwner: getEnvValue("", "GITHUB_REPOSITORY_OWNER").(string),
        RepoName:  getRepoName(),
        PRNumber:  getPRNumber(),
        BaseRef:   getEnvValue("", "GITHUB_BASE_REF").(string),
        HeadRef:   getEnvValue("", "GITHUB_HEAD_REF").(string),

        OllamaModel:       getEnvValue("qwen2:7b", "INPUT_OLLAMA-MODEL").(string),
        UseOllamaFallback: getEnvValue(true, "INPUT_USE-OLLAMA-FALLBACK").(bool),
    }

    if err := cfg.validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}

func (c *Config) validate() error {
    if c.GitHubToken == "" {
        return fmt.Errorf("github-token is required")
    }
    return nil
}
func getEnvValue(defaultValue interface{}, keys ...string) interface{} {
    var valueStr string
    for _, key := range keys {
        if v, exists := os.LookupEnv(key); exists {
            valueStr = v
            break
        }
    }

    if valueStr == "" {
        return defaultValue
    }

    // Parse the value based on the type of the default value.
    switch defaultValue.(type) {
    case string:
        return valueStr
    case int:
        if intValue, err := strconv.Atoi(valueStr); err == nil {
            return intValue
        }
    case float64:
        if floatValue, err := strconv.ParseFloat(valueStr, 64); err == nil {
            return floatValue
        }
    case bool:
        if boolValue, err := strconv.ParseBool(valueStr); err == nil {
            return boolValue
        }
    }

    // If parsing fails, return the default value.
    return defaultValue
}

func getRepoName() string {
    repo := os.Getenv("GITHUB_REPOSITORY")
    parts := strings.Split(repo, "/")
    if len(parts) >= 2 {
        return parts[1]
    }
    return ""
}

func getPRNumber() int {
    // GitHub Actions provides PR number differently
    if prNum := os.Getenv("GITHUB_EVENT_NUMBER"); prNum != "" {
        if num, err := strconv.Atoi(prNum); err == nil {
            return num
        }
    }
    
    // Parse from GITHUB_REF: refs/pull/123/merge
    if ref := os.Getenv("GITHUB_REF"); strings.Contains(ref, "refs/pull/") {
        parts := strings.Split(ref, "/")
        if len(parts) >= 3 {
            if num, err := strconv.Atoi(parts[2]); err == nil {
                return num
            }
        }
    }
    
    // Parse from GitHub event payload
    if eventPath := os.Getenv("GITHUB_EVENT_PATH"); eventPath != "" {
        if data, err := ioutil.ReadFile(eventPath); err == nil {
            var event struct {
                Number int `json:"number"`
                PullRequest struct {
                    Number int `json:"number"`
                } `json:"pull_request"`
            }
            if json.Unmarshal(data, &event) == nil {
                if event.Number > 0 {
                    return event.Number
                }
                if event.PullRequest.Number > 0 {
                    return event.PullRequest.Number
                }
            }
        }
    }
    
    return 0
}
