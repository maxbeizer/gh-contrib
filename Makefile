# Makefile for gh-contrib

.PHONY: build test help
.DEFAULT_GOAL := help

# Build the Go binary and reinstall the gh extension
## build: Build the Go binary and reinstall the gh extension
build:
	@echo "Building gh-contrib..."
	go build .
	@echo "Removing existing gh-contrib extension (if installed)..."
	gh ext remove gh-contrib || true # Ignore error if not installed
	@echo "Installing gh-contrib extension from current directory..."
	gh ext install .
	@echo "Build and reinstall complete."

# Run tests using the script/test runner
## test: Run the automated test suite
test:
	@echo "Running tests..."
	./script/test

# Show help for available make commands
## help: Show this help message
help:
	@echo "Available commands:"
	@grep -E '^## [a-zA-Z_-]+:.*' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ": "}; {gsub(/^## /, "", $$1); printf "  %-20s %s\n", $$1, $$2}'
