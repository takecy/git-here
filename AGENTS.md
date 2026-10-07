# Repository Guidelines

## Build, Test, and Development Commands
Before opening a PR, run `make build && make test && make lint` (`make lint` matches CI).

## Coding Style & Naming Conventions
Prefer the standard library over new dependencies when possible. Keep tests in the same package unless black-box coverage is required.
Repositories are processed in parallel, so every public `printer.Printer` method must hold `p.mu` while writing; cover new ones with a `*_ConcurrentSafe` test.
Inside `syncer.execute`, `eg.Go` callbacks must always `return nil`: record per-repo failures in `runStats` (→ exit 2) and return `error` from `Run` only for setup failures (→ exit 1).
This repo does not vendor dependencies; `make tidy` runs `go mod tidy` only — do not run `go mod vendor`.

## Testing Guidelines
This repository uses Go’s `testing` package plus `github.com/matryer/is`. Write focused tests with `t.Run(...)`; avoid table-driven tests here. Use `t.Parallel()` whenever the case is safe to parallelize. Name receiver-related tests as `Test{Struct}_Xxx`, for example `TestSync_Execute`.
Test `syncer` logic with `newSyncWithFake` (fake `Executor`) instead of spawning git. Real-git tests must be hermetic (bare repo under `t.TempDir()`, no network). Tests that swap `os.Stdout` cannot use `t.Parallel()`.

## Commit & Pull Request Guidelines
Follow Conventional Commits with English titles, scoped by package when helpful, for example `feat(gih): ...`, `refactor(syncer): ...`, and `fix(printer): ...`. PRs should follow `.github/PULL_REQUEST_TEMPLATE.md` with `## Issue` and `## Overview`, link the related issue, and describe behavioral changes briefly in Japanese. Include tests for functional changes and update `README.md` when CLI behavior changes.
Every push to `master` auto-tags a release (`.github/workflows/tagging.yml`). PRs are merge-committed, so each branch commit's type sets the bump: `fix` → patch, `feat` → minor, `BREAKING CHANGE:` footer → major, anything else → patch.

## Worktree Workflow
Do not work directly on the default branch (currently `master`). Start each non-trivial change from a dedicated branch and worktree under `.worktrees/`, for example `git worktree add .worktrees/feat-x -b feat/x`.
