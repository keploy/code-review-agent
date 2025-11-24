# KeployAI Code Review Action

[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Smart%20AI%20Code%20Review-blue.svg?colorA=24292e&colorB=0366d6&style=flat&longCache=true&logo=github)](https://github.com/marketplace/actions/smart-ai-code-review-action)

An intelligent GitHub Action that provides automated, AI-powered code reviews on your pull requests. It features a resilient fallback system, visual summaries, and smart diff analysis to deliver professional-grade feedback.


## Why This Is Helpful

This action improves code review workflow by:

- **Reducing reviewer workload** — AI handles routine comments so humans focus on critical architecture-level feedback.
- **Providing consistent feedback** — every PR receives structured, unbiased review.
- **Catching issues early** — flags style, performance, and potential bugs before human review.
- **Maintaining fast development velocity** — especially valuable for teams with frequent PRs.
- **Ensuring reliability** — fallback to Ollama guarantees reviews continue even if GitHub Models are unavailable.


## Key Features

-   📊 **Visual PR Summary**: Each review starts with a Mermaid sequence diagram summarizing the changes.
-   🧠 **Intelligent Diff Analysis**: Automatically prioritizes critical files in large pull requests to stay within token limits.
-   🔄 **Resilient Fallback System**: If the primary GitHub Models API fails, it seamlessly switches to a self-hosted Ollama model.
-   🤖 **Multi-Provider Support**: Natively supports GitHub Models and any model compatible with Ollama.
-   ✅ **Comprehensive Reviews**: Analyzes code for security, performance, best practices, and style.
-   🚀 **Zero-Setup Ready**: Works out of the box with the standard `GITHUB_TOKEN`.

## Usage

### Workflow Permissions

This action requires the following permissions in your workflow file to read code and post comments.

```yaml
permissions:
  contents: read
  pull-requests: write
```

### Basic Example

This configuration uses the default settings and is the quickest way to get started. Create a file like `.github/workflows/code-review.yml`:

```yaml
name: KeployAI Code Review

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
        uses: keploy/code-review-agent@v1
        with:
          github-token: ${{ secrets.GH_TOKEN }}
```

### Advanced Example

This configuration customizes the models, temperature, and file patterns.

```yaml
name: Advanced KeployAI Code Review

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
        uses: keploy/code-review-agent@v1
        with:
          github-token: ${{ secrets.GH_TOKEN }}
          model: 'gpt-4o'
          ollama-model: 'gemma3n:e4b'
          temperature: 0.1
          ignore-patterns: '*.md,*.lock,dist/**'
```

## Inputs 

| Name            | Required | Default      | Description                                   |
|-----------------|----------|--------------|-----------------------------------------------|
| `github-token`  | ✅       | –            | Token to read the repo and post PR comments.  |
| `model`         | ❌       | `gpt-4o`     | GitHub Models model used for the review.      |
| `ollama-model`  | ❌       | –            | Ollama model used as a fallback.             |
| `temperature`   | ❌       | `0.1`        | Lower = more deterministic reviews.          |
| `ignore-patterns` | ❌     | –            | Glob patterns to skip files from the review. |

See the [`action.yml`](action.yml) file for a full list of inputs and their descriptions.

## Contributing 🤝

We love contributions! Whether it’s improving documentation, fixing bugs, or adding new features — all PRs are welcome.

### How to Contribute

1. Fork the repo
2. Create a new branch (`feature/your-feature-name`)
3. Make your changes
4. Run tests (if applicable)
5. Submit a PR with a clear description of your changes

### Contribution Guidelines

- Keep PRs focused and scoped
- Use meaningful commit messages
- Describe the motivation behind the change
- If adding functionality, update the README when needed

### Need Help?

Open a discussion or issue — we’re happy to support contributors.


## License

This project is licensed under the [MIT License](LICENSE).
