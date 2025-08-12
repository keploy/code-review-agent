

#!/bin/bash

# Use underscore versions (cleaner)
export INPUT_GITHUB_TOKEN="your_github_token"
export GITHUB_REPOSITORY_OWNER="EuclidStellar"
export GITHUB_REPOSITORY="EuclidStellar/testtt"
export GITHUB_BASE_REF="pr_test_llm"
export GITHUB_HEAD_REF="pr_test"
export GITHUB_EVENT_NUMBER="126"

export INPUT_MODEL="gpt-4.1-nano"
export INPUT_MAX_TOKENS="6700"
export INPUT_TEMPERATURE="0.1"
export INPUT_IGNORE_PATTERNS="*.md,node_modules/**"
export INPUT_INCLUDE_PATTERNS=""

cd /Users/euclidstellar/Desktop/code-review-agent

echo "🚀 Testing Code Review Agent Locally"
echo "Repository: $GITHUB_REPOSITORY_OWNER/$GITHUB_REPOSITORY"
echo "PR #$GITHUB_EVENT_NUMBER: $GITHUB_BASE_REF → $GITHUB_HEAD_REF"
echo ""

# 🔧 Fix: Set up git remote and fetch branches
echo "🔍 Setting up git environment..."

# Check if we have the right remote
if ! git remote get-url origin 2>/dev/null | grep -q "EuclidStellar/testtt"; then
    echo "⚠️  Adding correct git remote..."
    git remote remove origin 2>/dev/null || true
    git remote add origin https://github.com/EuclidStellar/testtt.git
fi

# 🔧 FIX: Remove --all flag and fetch properly
echo "📥 Fetching remote branches..."
git fetch origin

# Check if branches exist
echo "🔍 Available branches:"
git branch -r | head -5

# Create local tracking branches if they don't exist
if ! git show-ref --verify --quiet refs/heads/pr_test_llm; then
    echo "🔧 Creating local branch pr_test_llm..."
    git checkout -b pr_test_llm origin/pr_test_llm 2>/dev/null || true
fi

if ! git show-ref --verify --quiet refs/heads/pr_test; then
    echo "🔧 Creating local branch pr_test..."
    git checkout -b pr_test origin/pr_test 2>/dev/null || true
fi

# Switch to a safe branch
git checkout main 2>/dev/null || git checkout master 2>/dev/null || git checkout pr_test_llm 2>/dev/null || true

echo ""
echo "✅ Git setup complete. Running code review agent..."
echo ""

go run main.go