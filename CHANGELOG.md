# Changelog

## v0.1.1
### Highlights
- Fixed lazy command initialization so commands only resolve dependencies they need, avoiding failures when OpenAI auth is unset for unrelated commands.
- Moved default CLI config storage from `~/.hange` to `~/.hange/config`.
- Refactored core domain packages under `domain/` and updated command wiring/factories accordingly.

### Changes
- Added: Lazy app builder accessors for auth, AI agent, git provider, and file provider dependencies.
- Changed: Internal package layout migrated from `pkg/` domain modules to `domain/`.
- Fixed: `make gen-mocks` cleanup and regenerated mocks/test imports to remove stale artifacts.
- Improved: Coverage and binary-size badge workflow outputs plus generated CLI command docs.
- Deleted: Legacy `pkg/factory/app.go`, `pkg/graceful/shutdown.go`, and obsolete `mocks/changes` mock set.

### Breaking Changes
- None.

### Known Issues
- None documented.

### Upgrade Notes
- Default config now uses `~/.hange/config`; re-run `hange auth "<token>"` or migrate your existing token into the new file.

## v0.1.0
### Highlights
- First public release of the `hange` CLI for generating commit messages and explaining code.

### Changes
- Added: Cobra-based CLI with `auth`, `explain`, `commit`, `commit-msg`, and `version` commands wired through factories and adapters.
- Added: OpenAI-backed processors for commit message generation and file/directory explanation with vector store support.
- Added: Config handling via Viper with env overrides and persisted OpenAI token storage.
- Added: Git adapter for staged diff intake and directory file streaming utilities.
- Added: Generated command docs in `docs/commands` and Make targets for coverage, mocks, and CLI docs.
- Changed: None.
- Fixed: None.
- Improved: None.
- Deleted: None.

### Breaking Changes
- None.

### Known Issues
- None documented.

### Upgrade Notes
- Install with `go install .` (or `go install github.com/yaroslav-koval/hange@latest`) and run `hange auth` to set your OpenAI token before using `commit`/`commit-msg`/`explain`.
