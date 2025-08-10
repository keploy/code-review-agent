package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"

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

    logger.Info("Configuration loaded: Model=%s, MaxTokens=%d", cfg.Model, cfg.MaxTokens)

    // Initialize clients
    ghClient := github.NewClient(cfg.GitHubToken, cfg.RepoOwner, cfg.RepoName)
    aiClient := reviewer.NewGitHubModelsClient(cfg.ModelsToken)
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

    // Generate AI review
    logger.Info("Generating AI review using %s", cfg.Model)
    review, err := aiClient.GenerateReview(prioritizedDiff, cfg.Model, cfg.Temperature, 4000)
    if err != nil {
        logger.Error("Failed to generate AI review: %v", err)
        // Post fallback review
        fallbackReview := generateFallbackReview(prioritizedDiff, err.Error())
        if postErr := ghClient.PostComment(ctx, cfg.PRNumber, fallbackReview); postErr != nil {
            log.Fatalf("Failed to post fallback review: %v", postErr)
        }
        logger.Info("Posted fallback review due to AI error")
        logger.GitHubOutput("review-posted", "true")
        return
    }

    // Post review to GitHub
    logger.Info("Posting review to PR #%d", cfg.PRNumber)
    if err := ghClient.PostComment(ctx, cfg.PRNumber, review); err != nil {
        log.Fatalf("Failed to post review: %v", err)
    }

    logger.Info("AI code review posted successfully")
    logger.GitHubOutput("review-posted", "true")
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
` + "```\n" + errorMsg + "\n```" + `

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