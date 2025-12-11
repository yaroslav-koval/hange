# Repository Guidelines

## Project Structure & Module Organization
- `main.go` boots the CLI; Cobra commands live in `cmd/` (auth, explain, version, root entrypoints with tests alongside).
- Core logic sits in `pkg/`: `agent/` orchestration, `auth/` domain + adapters, `config/` config loading, `consts/` shared constants, `crypt/` crypto helpers, `envs/` env resolution, `factory/` constructors. Keep new domain code inside `pkg/` with adapters beside interfaces.
- `mocks/` stores generated interfaces; `configs/badges/` holds coverage and binary size badge data.

## Build, Test, and Development Commands
- `go install . && hange -h` – build/install locally and verify commands are wired.
- `go test ./...` – fast test sweep for all packages.
- `make coverage` – run all tests with coverage report in `coverage.out` and human-readable summary.
- `make coverage-filtered` – coverage excluding `mocks` and `configs`.
- `make coverage-percent` – prints total coverage percentage only.
- `make gen-mocks` – regenerate mocks using `configs/.mockery.yml` (requires `mockery` on PATH).

## Coding Style & Naming Conventions
- Go 1.25+: run `gofmt`/`go fmt ./...` before committing; default tabs/column alignment.
- Package names are short lowercase; exported symbols use CamelCase; files follow Go’s `snake_words.go` pattern.
- Keep Cobra command files cohesive: flags/config in `cmd/`, business logic in `pkg/` functions for reuse.
- Always compare errors with `errors.Is`; never use direct equality.

## Testing Guidelines
- Tests use the Go test runner with `testify` assertions; place `_test.go` beside implementation files.
- Prefer table-driven cases for command behaviors and adapters. 
- Use mocks from `mocks/` for external interactions (set explicit `EXPECT()` calls); regenerate after interface changes.
- When a mock exists, use it instead of real implementations; assert expected calls (and non-calls) to document behavior.
- Keep coverage meaningful: run `make coverage-filtered` before PRs and ensure `coverage.out` remains up to date if referenced.

## Development Workflow
- Commit messages: concise, imperative summaries (e.g., `Add auth context token`, `Fix explain command flags`); annotate maintenance-only updates with `[skip ci]` if appropriate.
- PRs should include: brief description of the change, commands run (`go test ./...`, coverage target), and linked issues/tasks. Add screenshots or sample CLI output when altering user-facing behavior.
- Keep changesets focused; align new files with the existing layout (new commands in `cmd/`, shared logic in `pkg/`).
- Do not change existing code unless explicitly requested; prefer minimal, request-scoped diffs.
- When asked to write tests, avoid changing testable code; if alterations seem necessary, pause and confirm before modifying implementation.
