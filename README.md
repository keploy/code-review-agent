# Smart AI Code Review Action

[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Smart%20AI%20Code%20Review-blue.svg?colorA=24292e&colorB=0366d6&style=flat&longCache=true&logo=github)](https://github.com/marketplace/actions/smart-ai-code-review-action)

An intelligent GitHub Action that provides automated, AI-powered code reviews on your pull requests. It features a resilient fallback system, using **GitHub Models** as the primary provider and **Ollama** for local models as a backup, ensuring you always get a review.

## Key Features

-   🧠 **Intelligent Diff Analysis**: Automatically prioritizes critical files in large pull requests to stay within token limits.
-   🔄 **Resilient Fallback System**: If GitHub Models API fails or is rate-limited, it seamlessly switches to a self-hosted Ollama model.
-   🤖 **Multi-Provider Support**: Natively supports GitHub Models and any model compatible with Ollama.
-   ✅ **Comprehensive Reviews**: Analyzes code for security vulnerabilities, performance bottlenecks, best practices, and style.
-   ⚙️ **Highly Configurable**: Fine-tune models, ignore specific files, and adjust AI parameters.
-   🚀 **Zero-Setup Ready**: Works out of the box with the standard `GITHUB_TOKEN`.

## How It Works

1.  **GitHub Models First**: The action first attempts to generate a high-quality review using the specified GitHub Model (e.g., `gpt-4o-mini`).
2.  **Ollama Fallback**: If the primary API call fails (due to errors, rate limits, etc.), the action automatically installs Ollama inside the runner, pulls your specified model (e.g., `qwen2:7b`), and generates the review.
3.  **Static Fallback**: If both AI providers fail, a structured error report is posted, ensuring you are never left without feedback.

## Usage

Create a workflow file (e.g., `.github/workflows/code-review.yml`) in your repository.

### Basic Example

This configuration uses the default settings and is the quickest way to get started.

```yaml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read
  pull-requests: write

jobs:
  ai-review:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0 # Required to get the full git history for diffing

      - name: Run Smart AI Code Review
        uses: euclidstellar/code-review-agent@v1 # Replace with your repo name
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

### Advanced Example

This configuration customizes the models, temperature, and file patterns.

```yaml
name: Advanced AI Code Review

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read
  pull-requests: write

jobs:
  ai-review:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run Smart AI Code Review
        uses: euclidstellar/code-review-agent@v1 # Replace with your repo name
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          model: 'gpt-4o'
          ollama-model: 'llama3:8b'
          temperature: 0.5
          ignore-patterns: '*.md,*.lock,dist/**'
```

## Inputs

See the [`action.yml`](action.yml) file for a full list of inputs and their descriptions.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
