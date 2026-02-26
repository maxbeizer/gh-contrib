# Changelog

All notable changes to this project will be documented in this file.

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
