package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/euclidstellar/code-review-agent/internal/config"
	"github.com/euclidstellar/code-review-agent/internal/diff"
	"github.com/euclidstellar/code-review-agent/internal/github"
	"github.com/euclidstellar/code-review-agent/internal/reviewer"
	"github.com/euclidstellar/code-review-agent/internal/utils"
)

func main() {
	logger := utils.NewLogger()
	if err := run(logger); err != nil {
		logger.Error("Action failed: %v", err)
		os.Exit(1)
	}
	logger.Info("✅ Action completed successfully.")
}

func run(logger *utils.Logger) error {
	ctx := context.Background()
	logger.Info("Starting Smart AI Code Review Action v1.0.0")

	// Fix git ownership issue first
	setupGitSafeDirectory(logger)

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	logger.Info("Configuration loaded: Model=%s, MaxTokens=%d, OllamaFallback=%v",
		cfg.Model, cfg.MaxTokens, cfg.UseOllamaFallback)

	// Initialize clients
	ghClient := github.NewClient(cfg.GitHubToken, cfg.RepoOwner, cfg.RepoName)
	diffAnalyzer := diff.NewDiffAnalyzer(cfg.MaxTokens, cfg.IgnorePatterns, cfg.IncludePatterns)

	// Get and analyze PR diff
	logger.Info("Fetching PR diff for %s...%s", cfg.BaseRef, cfg.HeadRef)
	prDiff, err := diffAnalyzer.GetPRDiff(cfg.BaseRef, cfg.HeadRef)
	if err != nil {
		return fmt.Errorf("failed to get PR diff: %w", err)
	}
	if prDiff == "" {
		logger.Info("No changes found, skipping review.")
		logger.GitHubOutput("review-posted", "false")
		return nil
	}

	logger.Info("Analyzing diff (original size: %d bytes)", len(prDiff))
	prioritizedDiff, err := diffAnalyzer.AnalyzeAndPrioritize(prDiff, cfg.BaseRef, cfg.HeadRef)
	if err != nil {
		return fmt.Errorf("failed to analyze diff: %w", err)
	}
	logger.Info("Prioritized diff size: %d bytes (~%d tokens)", len(prioritizedDiff), len(prioritizedDiff)/4)

	// --- Review Logic ---
	review, provider, err := generateReview(ctx, cfg, prioritizedDiff, ghClient, logger)
	if err != nil {
		logger.Error("All review providers failed. Posting static fallback. Final error: %v", err)
		review = generateFallbackReview(prioritizedDiff, err.Error())
		provider = "static-fallback"
	}

	// Post the final review comment
	if err := ghClient.PostComment(ctx, cfg.PRNumber, review); err != nil {
		return fmt.Errorf("failed to post final comment: %w", err)
	}

	logger.Info("Posted review using provider: %s", provider)
	logger.GitHubOutput("review-posted", "true")
	logger.GitHubOutput("review-provider", provider)
	return nil
}

func generateReview(ctx context.Context, cfg *config.Config, diff string, ghClient *github.Client, logger *utils.Logger) (string, string, error) {
	// Attempt 1: GitHub Models
	logger.Info("🤖 Attempting review with GitHub Models (%s)...", cfg.Model)
	ghReview, err := tryGitHubModels(cfg, diff, logger)
	if err == nil {
		return ghReview, "github-models", nil
	}
	logger.Error("GitHub Models failed: %v", err)

	// Check if it's a token limit error
	if errors.Is(err, config.ErrTokenLimitExceeded) {
		// Post the friendly "coffee" message
		coffeeMessage := "Hey, it looks like your PR diff is very big, but don't worry, we got you! Grab a coffee, and before you finish it, your PR review will be ready. ☕"
		if postErr := ghClient.PostComment(ctx, cfg.PRNumber, coffeeMessage); postErr != nil {
			logger.Error("Failed to post 'coffee' comment: %v", postErr)
		}
	}

	// Attempt 2: Ollama Fallback
	if cfg.UseOllamaFallback {
		logger.Info("🔄 Attempting review with Ollama fallback (%s)...", cfg.OllamaModel)
		ollamaReview, err := tryOllamaFallback(ctx, cfg, diff, logger)
		if err == nil {
			return ollamaReview, "ollama", nil
		}
		logger.Error("Ollama fallback also failed: %v", err)
		return "", "", err // Return the last error
	}

	return "", "", err // Return the original error if Ollama is disabled
}

func tryGitHubModels(cfg *config.Config, diff string, logger *utils.Logger) (string, error) {
	logger.Info("🤖 Generating review with GitHub Models (%s)", cfg.Model)

	aiClient := reviewer.NewGitHubModelsClient(cfg.ModelsToken)
	review, err := aiClient.GenerateReview(diff, cfg.Model, cfg.Temperature, 4000)

	if err != nil {
		// Check for our specific token limit error
		if errors.Is(err, config.ErrTokenLimitExceeded) {
			return "", err // Pass the specific error up
		}
		// Check for other rate limit or quota errors
		if isRateLimitError(err) || isQuotaError(err) {
			return "", fmt.Errorf("rate limit/quota exceeded: %w", err)
		}
		return "", err
	}

	return review, nil
}

func tryOllamaFallback(ctx context.Context, cfg *config.Config, diff string, logger *utils.Logger) (string, error) {
	logger.Info("🦙 Setting up Ollama with model %s", cfg.OllamaModel)

	ollamaClient := reviewer.NewOllamaClient(cfg.OllamaModel)
	defer ollamaClient.Cleanup()

	// Setup Ollama (install, start service, pull model)
	if err := ollamaClient.SetupOllama(ctx); err != nil {
		return "", fmt.Errorf("failed to setup Ollama: %w", err)
	}

	// Generate review
	logger.Info("🔄 Generating review with Ollama...")
	review, err := ollamaClient.GenerateReview(ctx, diff)
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

func setupGitSafeDirectory(logger *utils.Logger) {
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