package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/euclidstellar/code-review-agent/internal/config"
	"github.com/euclidstellar/code-review-agent/internal/diff"
	"github.com/euclidstellar/code-review-agent/internal/github"
	"github.com/euclidstellar/code-review-agent/internal/reviewer"
	"github.com/euclidstellar/code-review-agent/internal/utils"
)

func main() {
	ctx := context.Background()
	logger := utils.NewLogger()

	logger.Info("Starting Smart AI Code Review Action v1.0.0")

	// Fix git ownership issue first
	setupGitSafeDirectory()

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Info("Configuration loaded: Model=%s, MaxTokens=%d, OllamaFallback=%v",
		cfg.Model, cfg.MaxTokens, cfg.UseOllamaFallback)

	// Initialize clients
	ghClient := github.NewClient(cfg.GitHubToken, cfg.RepoOwner, cfg.RepoName)
	diffAnalyzer := diff.NewDiffAnalyzer(cfg.MaxTokens, cfg.IgnorePatterns, cfg.IncludePatterns)

	// Get PR diff
	logger.Info("Fetching PR diff for %s...%s", cfg.BaseRef, cfg.HeadRef)
	prDiff, err := diffAnalyzer.GetPRDiff(cfg.BaseRef, cfg.HeadRef)
	if err != nil {
		log.Fatalf("Failed to get PR diff: %v", err)
	}

	if prDiff == "" {
		logger.Info("No changes found, skipping review")
		logger.GitHubOutput("review-posted", "false")
		return
	}

	// Analyze and prioritize diff
	logger.Info("Analyzing diff (original size: %d bytes)", len(prDiff))
	prioritizedDiff, err := diffAnalyzer.AnalyzeAndPrioritize(prDiff, cfg.BaseRef, cfg.HeadRef)
	if err != nil {
		log.Fatalf("Failed to analyze diff: %v", err)
	}

	logger.Info("Prioritized diff size: %d bytes (~%d tokens)", len(prioritizedDiff), len(prioritizedDiff)/4)

	// Try GitHub Models first
	review, err := tryGitHubModels(cfg, prioritizedDiff, logger)
	if err == nil {
		// Success with GitHub Models
		if err := ghClient.PostComment(ctx, cfg.PRNumber, review); err != nil {
			log.Fatalf("Failed to post review: %v", err)
		}
		logger.Info("✅ GitHub Models review posted successfully")
		logger.GitHubOutput("review-posted", "true")
		logger.GitHubOutput("review-provider", "github-models")
		return
	}

	// GitHub Models failed, try Ollama fallback
	logger.Error("GitHub Models failed: %v", err)

	if cfg.UseOllamaFallback {
		logger.Info("🔄 Trying Ollama fallback...")
		review, err := tryOllamaFallback(cfg, prioritizedDiff, logger)
		if err == nil {
			// Success with Ollama
			if err := ghClient.PostComment(ctx, cfg.PRNumber, review); err != nil {
				log.Fatalf("Failed to post Ollama review: %v", err)
			}
			logger.Info("✅ Ollama fallback review posted successfully")
			logger.GitHubOutput("review-posted", "true")
			logger.GitHubOutput("review-provider", "ollama")
			return
		}
		logger.Error("Ollama fallback also failed: %v", err)
	}

	// Both failed, post static fallback
	logger.Info("🔄 Using static fallback review...")
	fallbackReview := generateFallbackReview(prioritizedDiff, fmt.Sprintf("GitHub Models: %v, Ollama: %v", err, err))
	if err := ghClient.PostComment(ctx, cfg.PRNumber, fallbackReview); err != nil {
		log.Fatalf("Failed to post fallback review: %v", err)
	}
	logger.Info("Posted static fallback review")
	logger.GitHubOutput("review-posted", "true")
	logger.GitHubOutput("review-provider", "fallback")
}

func tryGitHubModels(cfg *config.Config, diff string, logger *utils.Logger) (string, error) {
	logger.Info("🤖 Generating review with GitHub Models (%s)", cfg.Model)

	aiClient := reviewer.NewGitHubModelsClient(cfg.ModelsToken)
	review, err := aiClient.GenerateReview(diff, cfg.Model, cfg.Temperature, 4000)

	if err != nil {
		// Check for rate limit or quota errors
		if isRateLimitError(err) || isQuotaError(err) {
			return "", fmt.Errorf("rate limit/quota exceeded: %w", err)
		}
		return "", err
	}

	return review, nil
}

func tryOllamaFallback(cfg *config.Config, diff string, logger *utils.Logger) (string, error) {
	logger.Info("🦙 Setting up Ollama with model %s", cfg.OllamaModel)

	ollamaClient := reviewer.NewOllamaClient(cfg.OllamaModel)
	defer ollamaClient.Cleanup()

	// Setup Ollama (install, start service, pull model)
	if err := ollamaClient.SetupOllama(); err != nil {
		return "", fmt.Errorf("failed to setup Ollama: %w", err)
	}

	// Generate review
	logger.Info("🔄 Generating review with Ollama...")
	review, err := ollamaClient.GenerateReview(diff)
	if err != nil {
		return "", fmt.Errorf("failed to generate Ollama review: %w", err)
	}

	// Add Ollama header to review
	reviewWithHeader := fmt.Sprintf("## 🦙 Ollama Code Review (%s)\n\n> **Note:** This review was generated using Ollama as a fallback when GitHub Models was unavailable.\n\n%s", cfg.OllamaModel, review)

	return reviewWithHeader, nil
}

func isRateLimitError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "429")
}

func isQuotaError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "quota") ||
		strings.Contains(errStr, "usage limit") ||
		strings.Contains(errStr, "insufficient")
}

func setupGitSafeDirectory() {
	// Configure git to trust the workspace directory
	commands := [][]string{
		{"git", "config", "--global", "--add", "safe.directory", "/github/workspace"},
		{"git", "config", "--global", "--add", "safe.directory", "*"},
	}

	for _, cmd := range commands {
		if output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Warning: Could not configure git safe directory: %s\n", string(output))
		}
	}
	fmt.Println("✅ Git safe directory configured")
}

func generateFallbackReview(diff, errorMsg string) string {
	return fmt.Sprintf(`## ❌ AI Code Review - Error Occurred

An error occurred while generating the automated code review.

### Error Details
`+"```\n"+errorMsg+"\n```"+`

### Manual Review Required
Please proceed with manual code review using these guidelines:

#### Quick Review Checklist
- **Security:** Check for vulnerabilities and exposed credentials
- **Performance:** Look for inefficient code patterns  
- **Testing:** Ensure adequate test coverage
- **Documentation:** Verify code is properly documented
- **Standards:** Confirm adherence to team coding standards

#### Diff Statistics
- **Size:** %d bytes (~%d tokens)
- **Requires manual analysis due to processing error**

---
**Error Time:** %s  
**Suggested Action:** Manual review and investigate workflow configuration`,
		len(diff), len(diff)/4, "now")
}