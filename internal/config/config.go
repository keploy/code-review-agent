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
}

func LoadFromEnv() (*Config, error) {
    cfg := &Config{
        // Try both hyphen and underscore versions
        GitHubToken:     getEnvWithFallback("INPUT_GITHUB-TOKEN", "INPUT_GITHUB_TOKEN", ""),
        ModelsToken:     getEnvWithFallback("INPUT_GITHUB-TOKEN", "INPUT_GITHUB_TOKEN", ""), // Same token
        Model:           getEnv("INPUT_MODEL", "gpt-4.1-nano"),
        MaxTokens:       getEnvIntWithFallback("INPUT_MAX-TOKENS", "INPUT_MAX_TOKENS", 6700),
        Temperature:     getEnvFloatWithFallback("INPUT_TEMPERATURE", "INPUT_TEMPERATURE", 0.1),
        IgnorePatterns:  strings.Split(getEnvWithFallback("INPUT_IGNORE-PATTERNS", "INPUT_IGNORE_PATTERNS", "*.md,node_modules/**"), ","),
        IncludePatterns: strings.Split(getEnvWithFallback("INPUT_INCLUDE-PATTERNS", "INPUT_INCLUDE_PATTERNS", ""), ","),
        RepoOwner:       getEnv("GITHUB_REPOSITORY_OWNER", ""),
        RepoName:        getRepoName(),
        PRNumber:        getPRNumber(),
        BaseRef:         getEnv("GITHUB_BASE_REF", ""),
        HeadRef:         getEnv("GITHUB_HEAD_REF", ""),
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

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
    if value := os.Getenv(key); value != "" {
        if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
            return floatValue
        }
    }
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

// Helper functions
func getEnvWithFallback(primary, fallback, defaultValue string) string {
    if value := os.Getenv(primary); value != "" {
        return value
    }
    if value := os.Getenv(fallback); value != "" {
        return value
    }
    return defaultValue
}

func getEnvIntWithFallback(primary, fallback string, defaultValue int) int {
    if value := os.Getenv(primary); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    if value := os.Getenv(fallback); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvFloatWithFallback(primary, fallback string, defaultValue float64) float64 {
    if value := os.Getenv(primary); value != "" {
        if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
            return floatValue
        }
    }
    if value := os.Getenv(fallback); value != "" {
        if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
            return floatValue
        }
    }
    return defaultValue
}