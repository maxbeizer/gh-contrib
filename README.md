# gh-contrib

`gh-contrib` is a GitHub CLI extension helping managers with managering tasks within the GitHub organization.

## Installation

```bash
gh ext install gh-contrib
```

## Usage

### List Pull Requests for a User

To list all pull requests created by a specific user in the GitHub organization:

```bash
gh contrib pulls <username>
```

Replace `<username>` with the GitHub username of the user whose pull requests you want to list.

### List Issues for a User

To list all issues created by a specific user in the GitHub organization:

```bash
gh contrib issues <username>
```

Replace `<username>` with the GitHub username of the user whose issues you want to list.

### List All Pull Requests and Issues for a User

To list all pull requests and issues created by a specific user in the GitHub organization:

```bash
gh contrib all <username>
```

Replace `<username>` with the GitHub username of the user whose pull requests and issues you want to list.

### Summarize Pull Request or Issue Bodies

To summarize pull request or issue bodies passed via stdin, separated by the delimiter `---END-OF-ENTRY---`:

```bash
gh contrib summarize
```

This command processes each entry individually and provides a summary in bullet points.

### Debug Mode

To enable debug mode and see additional information during execution, use the `--debug` flag:

```bash
gh contrib --debug pulls <username>
```

## Example

```bash
gh contrib pulls octocat
```

This will display all pull requests created by the user `octocat` in the GitHub organization, sorted by the most recently created.

## Requirements

- Go 1.16 or later
- GitHub CLI installed and authenticated

## Flags

### `--since`

Use the `--since` flag to filter pull requests created since a specific date. The date should be in the format `YYYY-MM-DD`. If not provided, it defaults to 30 days before the current date.

Example:

```bash
gh contrib --since 2025-04-01 pulls <username>
```

### `--body-only`

Use the `--body-only` flag to fetch and print only the body of the pull requests/issues.

Example:

```bash
gh contrib --body-only pulls <username>
```

### `--org`

Use the `--org` flag to override the configured organization for a specific command. This is useful if you want to temporarily query a different organization without changing the configuration.

Example:

```bash
gh contrib --org primer pulls <username>
```

This will fetch pull requests authored by `<username>` in the `primer` organization, regardless of the configured organization.

> [!NOTE]
> The search API does not currently support `OR` queries so as of this writing you can only query one org at a time :(

### `--model`

Use the `--model` flag to override the AI model used for summarization. This will override the `model` key in the configuration or the default `gpt-4o`.

Example:

```bash
# Override the model when summarizing entries
gh contrib --model gpt-3.5 summarize
```

See this page of list of available models https://learn.microsoft.com/en-us/azure/ai-services/openai/concepts/models

## Configuration

To configure the organization or model used by `gh-contrib`, update the `~/.config/gh/config.yml` file under the `extensions` block for the `gh-contrib` extension. For example:

```yaml
extensions:
  gh-contrib:
    org: my-custom-org
    model: gpt-4o
```

- Replace `my-custom-org` with the desired organization name. If the `org` key is not set, the tool defaults to using `github` as the organization.
- Replace `gpt-4o` with the desired model name. If the `model` key is not set, the tool defaults to using `gpt-4o` as the model.

## Testing

To run the automated test suite (including race condition detection and timeouts), execute the `script/test` script:

```bash
./script/test
```

## Development

This project includes a `Makefile` to streamline common development tasks.

- **`make build`**: Compiles the Go binary and reinstalls the `gh` extension from the local source. This is useful for testing changes quickly.
- **`make test`**: Runs the automated test suite using the `./script/test` runner.
- **`make help`** or **`make`**: Displays a list of available `make` commands and their descriptions.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
