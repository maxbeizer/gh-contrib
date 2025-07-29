# 📊 gh-contrib

> **A powerful GitHub CLI extension to visualize and understand contributions across your organization**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.16+-blue.svg)](https://golang.org/)

`gh-contrib` helps you track, visualize, and analyze GitHub contributions with beautiful graphs, detailed summaries, and AI-powered insights. Perfect for team leads, project managers, and developers who want to understand contribution patterns.

## ✨ Features

- 📈 **Visual contribution graphs** - See weekly activity patterns at a glance
- 🔍 **Detailed contribution lists** - Pull requests, issues, and combined views
- 🤖 **AI-powered summaries** - Automatically summarize PR/issue content
- 🎯 **Flexible filtering** - Filter by date ranges and organizations
- ⚡ **Fast and intuitive** - Built with Go for speed and efficiency

## 🚀 Quick Start

### Installation

```bash
gh ext install maxbeizer/gh-contrib
```

### Your First Command

```bash
# See your own contributions from the last 30 days
gh contrib graph

# Or check someone else's contributions
gh contrib graph octocat
```

> 💡 **Tip:** If no username is provided, the extension automatically uses your GitHub username!

## 📖 Usage Guide

### 📊 Visualize Contributions

Create beautiful weekly contribution graphs:

```bash
gh contrib graph [username]
```

**Example output:**

```
Week  1 (Apr 15 - Apr 21): •□■
Week  2 (Apr 22 - Apr 28): ○••
Week  3 (Apr 29 - May 05): ■□

Legend:
• = Closed PR  ○ = Open PR  ■ = Closed Issue  □ = Open Issue

Total Contributions: 7 over 31 days (avg: 0.23 per day)
PRs: 4 total (3 closed, 1 open)
Issues: 3 total (1 closed, 2 open)
```

### 🔍 List Contributions

**Pull Requests Only:**

```bash
gh contrib pulls [username]
```

**Issues Only:**

```bash
gh contrib issues [username]
```

**Everything Together:**

```bash
gh contrib all [username]
```

### 🤖 AI-Powered Summaries

Summarize multiple PR/issue descriptions using AI:

```bash
gh contrib summarize
```

Pass content via stdin, separated by `---END-OF-ENTRY---` delimiters.

### 🐛 Debug Mode

Get detailed execution information:

```bash
gh contrib --debug graph octocat
```

## 🎛️ Advanced Options

### 📅 Date Filtering

Filter contributions by date range:

```bash
# Get contributions since a specific date
gh contrib --since 2025-04-01 pulls octocat

# Works with all commands
gh contrib --since 2025-04-01 graph octocat
```

**Date format:** `YYYY-MM-DD` (defaults to 30 days ago if not specified)

### 📝 Content Focus

Get just the content without metadata:

```bash
# Show only PR/issue body content
gh contrib --body-only pulls octocat
```

### 🏢 Organization Override

Query different organizations on the fly:

```bash
# Check contributions in a specific org
gh contrib --org primer pulls octocat
```

> ⚠️ **Note:** GitHub's search API doesn't support OR queries, so you can only query one organization at a time.

### 🤖 AI Model Selection

Choose your preferred AI model for summaries:

```bash
# Use a specific model for summarization
gh contrib --model gpt-3.5 summarize
```

[View available models →](https://learn.microsoft.com/en-us/azure/ai-services/openai/concepts/models)

## ⚙️ Configuration

Customize default settings in `~/.config/gh/config.yml`:

```yaml
extensions:
  gh-contrib:
    org: my-custom-org # Default organization
    model: gpt-4o # Default AI model
```

**Configuration options:**

- `org`: Default organization name (fallback: `github`)
- `model`: Default AI model (fallback: `gpt-4o`)

## 🛠️ Development & Testing

### Prerequisites

- Go 1.16 or later
- GitHub CLI installed and authenticated

### Quick Development

```bash
# Build and test locally
make build

# Run tests
make test

# See all available commands
make help
```

### Testing

Run the comprehensive test suite:

```bash
./script/test
```

Includes race condition detection and timeout handling.

## 💡 Pro Tips

- **🔄 Default behavior:** All commands default to your own username when none is provided
- **📊 Best visualization:** Use `graph` command for quick visual insights
- **🎯 Focused analysis:** Combine `--since` with specific date ranges for targeted analysis
- **🏃‍♂️ Quick debugging:** Add `--debug` to any command for detailed execution info

## 📋 Examples

```bash
# Quick personal overview
gh contrib graph

# Team member analysis (last 2 weeks)
gh contrib --since 2025-07-15 all teammate

# Organization-specific search
gh contrib --org microsoft pulls octocat

# Debug a slow query
gh contrib --debug --since 2025-01-01 graph octocat
```

## 📜 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for the GitHub community**

[Report Bug](https://github.com/maxbeizer/gh-contrib/issues) • [Request Feature](https://github.com/maxbeizer/gh-contrib/issues) • [Contribute](https://github.com/maxbeizer/gh-contrib/pulls)

</div>
