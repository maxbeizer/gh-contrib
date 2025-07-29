# GitHub Copilot Instructions for gh-contrib

This document provides guidelines for GitHub Copilot when writing code for the `gh-contrib` project. Follow these patterns to maintain consistency, security, and maintainability.

## 🏗️ Architecture Patterns

### Dependency Injection

- **Always use interfaces** for external dependencies (GitHub API, HTTP clients, AI services)
- **Define interfaces first** before concrete implementations
- **Group interfaces at the top** of files with clear documentation

```go
// GitHubClient defines the methods needed to interact with the GitHub API.
type GitHubClient interface {
    Get(path string, response interface{}) error
}

// TokenFetcher defines the method needed to fetch an authentication token.
type TokenFetcher interface {
    FetchToken() (string, error)
}
```

### Constructor Pattern

- **Use constructor functions** for complex types that require initialization
- **Return errors** from constructors when initialization can fail
- **Validate parameters** in constructors

```go
func NewDefaultGitHubClient() (*DefaultGitHubClient, error) {
    client, err := api.DefaultRESTClient()
    if err != nil {
        return nil, fmt.Errorf("error creating default GitHub API client: %w", err)
    }
    return &DefaultGitHubClient{client: client}, nil
}
```

## 🔒 Security Best Practices

### Error Handling

- **Always wrap errors** with context using `fmt.Errorf` and `%w` verb
- **Never expose sensitive information** in error messages
- **Validate all external inputs** before processing

```go
if err != nil {
    return "", fmt.Errorf("error creating default GitHub API client: %w", err)
}
```

### Token Management

- **Never log or print tokens** in debug output
- **Use dedicated token fetcher interfaces** rather than hardcoded tokens
- **Handle token fetch failures gracefully**

```go
token, err := s.tokenFetcher.FetchToken()
if err != nil {
    return "", fmt.Errorf("error retrieving GitHub token: %w", err)
}
// Never log the token value
```

### Input Validation

- **Sanitize all user inputs** before using in API calls
- **Use URL encoding** for query parameters
- **Validate date formats** before parsing

```go
// Always URL encode user inputs
query := fmt.Sprintf("author:%s %s", url.QueryEscape(login), baseQuery)
```

## 🧪 Testing Standards

### Mock Implementation

- **Create comprehensive mocks** for all external dependencies
- **Record method calls** for verification in tests
- **Allow customizable behavior** through function fields

```go
type MockGitHubClient struct {
    GetFunc  func(path string, response interface{}) error
    GetCalls []string
}

func (m *MockGitHubClient) Get(path string, response interface{}) error {
    m.GetCalls = append(m.GetCalls, path)
    if m.GetFunc != nil {
        return m.GetFunc(path, response)
    }
    return nil
}
```

### Test Structure

- **Use table-driven tests** for multiple scenarios
- **Capture output** for CLI commands testing
- **Reset global state** between tests
- **Test error conditions** explicitly

```go
func captureOutput(f func()) (string, string) {
    // Capture stdout and stderr
    // Return both for verification
}

func resetFlags() {
    debug = false
    since = time.Now().AddDate(0, 0, -30).Format(dateFormat)
    bodyOnly = false
}
```

## 📊 Data Handling

### JSON Parsing

- **Use struct tags** for JSON field mapping
- **Handle optional fields** with `omitempty`
- **Validate parsed data** before use

```go
type GitHubItem struct {
    Number     int    `json:"number"`
    Title      string `json:"title"`
    HTMLURL    string `json:"html_url"`
    Body       string `json:"body,omitempty"`
    CreatedAt  string `json:"created_at"`
}
```

### Time Handling

- **Use consistent date formats** (define constants)
- **Parse dates properly** with proper error handling
- **Consider timezone implications** in date calculations

```go
const dateFormat = "2006-01-02"

createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
if err != nil {
    return fmt.Errorf("error parsing created date: %w", err)
}
```

## 🔧 Code Organization

### Global Variables

- **Minimize global state** - prefer dependency injection
- **Group related globals** together with clear comments
- **Initialize globals** in `init()` functions when necessary

```go
// Global variables for command-line flags
var (
    debug     bool
    since     string
    bodyOnly  bool
    orgFlag   string
    modelFlag string
)
```

### Function Design

- **Keep functions focused** on single responsibilities
- **Use descriptive names** that explain the function's purpose
- **Limit function parameters** (prefer structs for complex parameter sets)
- **Return errors as the last return value**

```go
func handlePullsCommand(args []string, client GitHubClient) {
    // Focused on handling pulls command only
}

func fetchAllResults(client GitHubClient, searchURL string) ([]GitHubItem, error) {
    // Clear purpose and error handling
}
```

## 🎯 Performance Considerations

### Memory Management

- **Reuse buffers** when possible
- **Close resources** properly (HTTP response bodies, files)
- **Avoid memory leaks** in long-running operations

```go
defer resp.Body.Close()

// Reuse writers
writer := csv.NewWriter(os.Stdout)
defer writer.Flush()
```

### API Efficiency

- **Implement pagination** for large result sets
- **Rate limit API calls** when necessary
- **Cache results** when appropriate

```go
func fetchAllResults(client GitHubClient, searchURL string) ([]GitHubItem, error) {
    var allItems []GitHubItem
    page := 1

    for {
        // Implement proper pagination
        // Handle rate limits
        // Accumulate results efficiently
    }
}
```

## 📝 Documentation

### Code Comments

- **Document all public interfaces** and types
- **Explain complex business logic** with inline comments
- **Use godoc-style comments** for package documentation

```go
// GitHubClient defines the methods needed to interact with the GitHub API.
// Implementations should handle authentication and rate limiting appropriately.
type GitHubClient interface {
    Get(path string, response interface{}) error
}
```

### Error Messages

- **Provide actionable error messages** to users
- **Include context** about what operation failed
- **Suggest solutions** when possible

```go
fmt.Printf("No issues found for user '%s' in the '%s' organization.\n", login, org)
```

## 🛠️ Development Workflow

### Flag Handling

- **Support flexible flag positioning** in commands
- **Provide sensible defaults** for all flags
- **Validate flag combinations** for conflicts

```go
// Support flags anywhere in command line
var nonFlagArgs []string
for i < len(args) {
    if strings.HasPrefix(arg, "-") {
        // Handle flag extraction
    } else {
        nonFlagArgs = append(nonFlagArgs, arg)
    }
}
```

### Build Integration

- **Use Makefiles** for common development tasks
- **Provide clear targets** with descriptions
- **Include testing** in build pipeline

```makefile
## build: Build the Go binary and reinstall the gh extension
build:
    go build .
    gh ext remove gh-contrib || true
    gh ext install .
```

## 🚨 Common Pitfalls to Avoid

1. **Don't log sensitive data** (tokens, user data)
2. **Don't ignore errors** - always handle or propagate them
3. **Don't use global state** for test-dependent values
4. **Don't hardcode URLs or magic numbers** - use constants
5. **Don't forget to close resources** (HTTP bodies, files)
6. **Don't assume API responses** are always valid - validate first
7. **Don't couple code tightly** - use interfaces for testability

## 🎨 Code Style

- **Use gofmt** for consistent formatting
- **Follow Go naming conventions** (PascalCase for exports, camelCase for private)
- **Group imports** logically (standard library, external, internal)
- **Use meaningful variable names** over short abbreviations
- **Prefer explicit error handling** over ignoring errors

## 🔄 Maintenance Guidelines

- **Write tests first** for new features when possible
- **Update documentation** when changing interfaces
- **Consider backward compatibility** for configuration changes
- **Use semantic versioning** principles for releases
- **Review security implications** of all changes

---

By following these guidelines, ensure that all code contributions maintain the high standards of security, testability, and maintainability that characterize the `gh-contrib` project.
