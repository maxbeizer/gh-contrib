# AI Assistant Context for gh-contrib

This document provides context for AI assistants working with the `gh-contrib` codebase.

## Project Overview

`gh-contrib` is a command-line tool written in Go. Its primary purpose is to interact with the GitHub API to fetch information about pull requests (PRs) and issues associated with a specific user within the `github` organization. It can output this information in CSV format or extract just the bodies of the items. Additionally, it features a `summarize` command that leverages an external AI model (via Azure Inference AI) to summarize the content of PR and issue bodies.

## Codebase Structure

The project currently consists of a single main file:

- `main.go`: Contains all the core logic, including command parsing, flag handling, API interactions (GitHub and AI), data processing, and output formatting.
- `go.mod` / `go.sum`: Manage Go module dependencies.
- `README.md`: General project documentation (likely).

## Key Functionality & Commands

- **Fetching Data:** Uses the `github.com/cli/go-gh/v2/pkg/api` library to interact with the GitHub REST API (specifically the `search/issues` endpoint). It handles pagination to retrieve all relevant results.
- **Commands:**
  - `pulls <login>`: Fetches PRs authored by `<login>` in the `github` org.
  - `issues <login>`: Fetches issues authored by `<login>` in the `github` org.
  - `all <login>`: Fetches both PRs and issues authored by `<login>` in the `github` org.
  - `summarize`: Reads text (expected to be PR/issue bodies separated by `---END-OF-ENTRY---`) from stdin or an argument and sends each entry to an AI endpoint for summarization.
- **Flags:**
  - `--since YYYY-MM-DD`: Filters results created after the specified date.
  - `--body-only`: Outputs only the bodies of the fetched items, formatted with start/end markers, instead of CSV. Useful for piping to the `summarize` command.
  - `--debug`: Enables verbose logging.
- **Authentication:**
  - GitHub API: Relies on the user being authenticated via the `gh` CLI, as `go-gh` uses this context.
  - AI Summarization: Explicitly retrieves a token using `gh auth status --show-token` and sends it as a Bearer token to the Azure Inference AI endpoint (`https://models.inference.ai.azure.com/chat/completions`).

## Development Guidelines

### Formatting

- Follow standard Go formatting practices. Use `gofmt` or `goimports` to format the code before committing.
- Use the defined constants (e.g., `dateFormat`, `defaultModel`, `aiEndpoint`, `entryDelimiter`) instead of magic strings.

### Readability

- Use clear and descriptive names for variables, functions, and types.
- Keep functions focused on a single responsibility.
- Add comments to explain complex logic or non-obvious code sections.

### Dependencies

- Manage Go dependencies using Go modules (`go.mod`, `go.sum`).
- The tool has a runtime dependency on the `gh` CLI being installed and configured for the `summarize` command's authentication mechanism.

### Testing

- Automated tests are provided in `main_test.go`, using dependency injection with mock implementations of `GitHubClient`, `TokenFetcher`, and `Summarizer`.
- Tests capture stdout and stderr to verify CLI output and simulate API responses without making real network calls.
- Run the full test suite with:
  ```shell
  go test ./...
  ```
- When adding new functionality or fixing bugs, extend the test coverage with unit or integration tests using the same mocking patterns.

### Running the Tool

1.  Ensure Go and the `gh` CLI are installed.
2.  Authenticate with GitHub using `gh auth login`.
3.  Build the tool: `go build .`
4.  Run the executable: `./gh-contrib <command> [args...] [flags...]`
    - Example: `./gh-contrib pulls octocat --since 2025-04-01`
    - Example: `./gh-contrib all octocat --body-only | ./gh-contrib summarize`

### Error Handling

- Check for errors returned by function calls, especially API interactions and I/O operations.
- Provide informative error messages to the user via `fmt.Println`. Consider using `fmt.Fprintf(os.Stderr, ...)` for errors to distinguish them from standard output.

### API Interaction

- Be mindful of GitHub API rate limits. The `fetchAllResults` function handles pagination correctly.
- The AI summarization endpoint and model (`gpt-4o` via Azure) are hardcoded. Changes might require updating the `aiEndpoint` and `defaultModel` constants and potentially the payload structure.
