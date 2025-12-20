# Changelog

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
