# Smart AI Code Review Action

AI-powered code review with **intelligent fallback system**:
1. **GitHub Models** (primary, fastest)
2. **Ollama** (fallback for rate limits)
3. **Static review** (final fallback)

## Features

✅ **Smart Diff Prioritization** - Focuses on critical files  
✅ **Multi-Model Support** - GitHub Models + Ollama fallback  
✅ **Rate Limit Resilience** - Never fails due to API limits  
✅ **Comprehensive Reviews** - Security, performance, best practices  
✅ **Zero Configuration** - Works out of the box  

## Usage

```yaml
- uses: euclidstellar/code-review-agent@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    model: 'gpt-4o-mini'
    ollama-model: 'qwen2.5:7b'  # Fallback model
```

## Advanced Configuration

```yaml
- uses: euclidstellar/code-review-agent@v1
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    model: 'gpt-4o-mini'
    max-tokens: 6700
    temperature: 0.1
    enable-ollama: true
    ollama-model: 'qwen2.5:7b'
    use-ollama-fallback: true
    ignore-patterns: '*.md,node_modules/**'
```
