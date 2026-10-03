# Repository Guidelines

## Build, Test, and Development Commands
Before opening a PR, run `make build && make test && make lint` (`make lint` matches CI).

## Coding Style & Naming Conventions
Prefer the standard library over new dependencies when possible. Keep tests in the same package unless black-box coverage is required.
Repositories are processed in parallel, so every public `printer.Printer` method must hold `p.mu` while writing; cover new ones with a `*_ConcurrentSafe` test.

## Testing Guidelines
This repository uses Go’s `testing` package plus `github.com/matryer/is`. Write focused tests with `t.Run(...)`; avoid table-driven tests here. Use `t.Parallel()` whenever the case is safe to parallelize. Name receiver-related tests as `Test{Struct}_Xxx`, for example `TestSync_Execute`.

## Commit & Pull Request Guidelines
Follow Conventional Commits with English titles, scoped by package when helpful, for example `feat(gih): ...`, `refactor(syncer): ...`, and `fix(printer): ...`. PRs should follow `.github/PULL_REQUEST_TEMPLATE.md` with `## Issue` and `## Overview`, link the related issue, and describe behavioral changes briefly in Japanese. Include tests for functional changes and update `README.md` when CLI behavior changes.

## Worktree Workflow
Do not work directly on the default branch (currently `master`). Start each non-trivial change from a dedicated branch and worktree under `.worktrees/`, for example `git worktree add .worktrees/feat-x -b feat/x`.
