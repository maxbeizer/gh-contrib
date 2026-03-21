# Copilot Instructions for gh-extension-template

## Project Overview
This is a Go-based GitHub CLI extension that demonstrates best practices for building extensible CLI tools. The project uses the `gh` CLI extension framework and follows modern Go development patterns.

## Technology Stack
- **Language**: Go
- **Framework**: GitHub CLI (gh) extension framework  
- **Testing**: Go standard library testing package

## Go Development Guidelines

### Code Organization
- Follow Go project layout conventions (main package in root, reusable code in subdirectories)
- Use the `go mod` system for dependency management
- Organize packages by functionality, not by layer

### Go Idioms and Standard Library
- Use the Go standard library exclusively where possible (avoid external dependencies)
- Follow idiomatic Go patterns:
  - Use interfaces for abstraction and testability
  - Prefer composition over inheritance
  - Use goroutines and channels for concurrency
  - Handle errors explicitly, no exceptions
  - Use defer for resource cleanup
- Return errors as the last return value: `(result Type, err error)`

### Naming Conventions
- **Variables and Functions**: Use camelCase for unexported, PascalCase for exported
- **Constants**: Use ALL_CAPS for package-scoped constants
- **Packages**: Use short, lowercase names (single word where possible)
- **Receivers**: Use short 1-2 letter names like `f`, `s`, `r` for receiver variables
- **Methods**: Name methods by their action and what they return (e.g., `GetConfig()`, `IsValid()`)

### Testing
- Write table-driven tests for comprehensive coverage:
  - Define test cases as a slice of anonymous structs
  - Include input, expected output, and error conditions
  - Use subtests with `t.Run()` for better test isolation and reporting
- Test file naming: `*_test.go` in the same package
- Aim for high coverage, especially for CLI argument parsing and core logic

### Example Table-Driven Test Pattern
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "example", "output", false},
        {"invalid input", "", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## GitHub CLI Extension Patterns

### Key Extension Framework Concepts
- Use `github.com/cli/go-gh` for interacting with the GitHub API
- Leverage the `cmdutil` package for common CLI patterns
- Use `flag.FlagSet` for command-line argument parsing
- Follow the standard Go CLI pattern: `command -> subcommand -> flags`

### Command Structure
- Create a main command that handles core functionality
- Use subcommands for different operational modes
- Support `--help` and `-h` flags automatically
- Provide clear, concise error messages for users
- Use exit codes appropriately (0 for success, 1 for general errors)

### Configuration and Flags
- Support both flags and configuration files when appropriate
- Use environment variables for sensitive data (following gh CLI conventions)
- Validate all inputs before processing
- Provide sensible defaults

## Development Workflow

### Building and Running
- Build: `go build -o gh-extension-template ./cmd/...` or use Makefile
- Run: `./gh-extension-template` or through gh CLI: `gh extension-template <command>`
- Test: `go test ./...`
- Coverage: `go test -cover ./...`

### Code Quality
- Run gofmt before committing: `gofmt -w .`
- Use golint or golangci-lint for linting
- Ensure all tests pass
- Keep functions focused and small (aim for < 50 lines)
- Write clear comments for exported functions and types

## Documentation
- Document exported functions and types with comment strings
- Include examples in function documentation where helpful
- Update README.md for user-facing changes
- Document command-line flags and their usage
