# Repository Guidelines

## Project Structure & Module Organization
`gih/main.go` is the CLI entry point. Core behavior lives in `syncer/`: `dir.go` discovers direct child repositories, `git.go` runs git commands, and `syncer.go` coordinates parallel execution. Output formatting lives in `printer/`. Tests sit next to the code as `*_test.go`. CI and templates are under `.github/`.

## Build, Test, and Development Commands
Use Go 1.26.

- `make build`: builds a development binary as `./gih_dev`.
- `make install`: installs `gih` from `./gih`.
- `make test`: runs `go test -v -race ./...`.
- `make lint`: runs `golangci-lint` as used in CI.
- `make tidy`: normalizes `go.mod` and `go.sum`.
- `DEBUG=* go run ./gih status`: handy for manual CLI checks.

Before opening a PR, run `make build && make test && make lint`.

## Coding Style & Naming Conventions
Follow standard Go formatting with `gofmt`; keep packages small and focused. Prefer the standard library over new dependencies when possible. Use clear package scopes such as `feat(syncer)` or `fix(printer)` when naming changes. Keep filenames lowercase, and keep tests in the same package unless black-box coverage is required.

## Testing Guidelines
This repository uses Go’s `testing` package plus `github.com/matryer/is`. Write focused tests with `t.Run(...)`; avoid table-driven tests here. Use `t.Parallel()` whenever the case is safe to parallelize. Name receiver-related tests as `Test{Struct}_Xxx`, for example `TestSync_Execute`. Run targeted checks with `go test -v ./syncer/` and use `go test -cover ./...` when touching behavior broadly.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commits, for example `feat(gih): ...`, `refactor(syncer): ...`, and `fix(printer): ...`. Keep commit titles in English and scoped when helpful. PRs should follow `.github/PULL_REQUEST_TEMPLATE.md` with `## Issue` and `## Overview`, link the related issue, and describe behavioral changes briefly in Japanese. Include tests for functional changes and update `README.md` when CLI behavior changes.

## Worktree Workflow
Do not work directly on `master`. Start each non-trivial change from a dedicated branch and worktree under `.worktrees/`, for example `git worktree add .worktrees/feat-x -b feat/x`.
