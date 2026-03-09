# Changelog

All notable changes to this project will be documented in this file.

## 0.7.0 - 2026-03-09

- Add `--visibility` flag to filter contributions by repository visibility (`public` or `private`)
- Add CI workflow for running tests on push and PR
- Add `CODE_OF_CONDUCT.md`
- Add `ci` Makefile target
- Expand `.gitignore` with IDE and build artifact patterns

## 0.6.0 - 2026-02-26

- Fetch PRs, reviews, issues, and discussions concurrently in `all` and `graph` commands (~4x faster)
- Thread-safe internals for concurrent API calls

## 0.5.1 - 2026-02-26

- Fix discussions query returning empty results (removed erroneous `type:discussion` from GraphQL search string)
- Default to the authenticated user when no `<login>` argument is provided for all commands
- Extract shared `resolveLogin()` helper for consistent behavior across commands

## 0.5.0 - 2026-02-26

- Add `discussions` command to capture discussions authored by a user via GraphQL API
- Add `reviews` command to capture PRs reviewed by a user
- Add `--prompt-only` flag for agentic workflow composability with `summarize`
- Include reviews and discussions in `all` and `graph` commands
- Graph uses ◆/◇ for reviews and ▲/△ for discussions

## 0.3.0 - 2025-07-01

- Default to the current user if no username is passed in
