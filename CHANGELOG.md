# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.2.0] - 2026-03-15

### Added

- TOML-configurable fuzzy filter for interactive TUI (`[tui] filter = true`).
  - Real-time substring filtering by slot name or branch name.
  - `ctrl+j/k` navigation in filter mode (j/k used for typing).
- GoReleaser integration for automated multi-platform releases.
- GitHub Actions release workflow triggered on `v*` tag push.
- Homebrew tap support via `AquiTCD/homebrew-tap`.
- `/bump-version` agent workflow for automated release process.

### Changed

- Color output is now always enabled by default. Only `NO_COLOR` environment variable disables it.
  - Fixes color not displaying when using the `gsl` wrapper.
- README comprehensively updated: Homebrew install, config examples, command flags, tech stack, gwq acknowledgement.

### Fixed

- Duplicate error messages in CLI output.
- Missing `-b` alias for `--branch` flag.
- Error display for unknown slot and branch-in-use scenarios.

## [0.1.1] - 2026-03-14

### Added

- `-g` shorthand for `--global` flag in `git slot --init`.
- Proactive `PATH` environment check post-install in `Makefile`.

### Fixed

- Fix `gsl` shell wrapper double execution bug when command fails or returns result.
- Improve `make install` to correctly handle `GOBIN` and `GOPATH` targets.

## [0.1.0] - 2026-03-13

- TOML-based configuration system with global (`~/.config/git-slot/config.toml`) and project (`git-slot.toml`) hierarchy
- Slot management: Load, Clear, List, GetPath, Swap, Status operations
- `git slot <slot> <branch>` to load an existing branch into a slot
- `git slot <slot> -c <branch>` to create a new branch and load it
- `git slot <slot>` to print the slot's worktree path
- `git slot --list` to display all slots and their status
- `git slot --clear <slot>` to remove a slot's worktree
- `git slot --swap <A> <B>` to swap branches between two slots
- `git slot --status [slot]` for detailed slot status
- `git slot --init` to generate a template configuration file
- `git slot --version` to display version information
- `--json` flag for machine-readable output on `--list` and `--status`
- `--force` flag to skip safety checks on dirty worktrees
- Safety guards: branch duplication detection, dirty state protection
- gwq-compatible directory structure (`~/worktrees/{host}/{owner}/{repo}/slots/`)
- Shell completion for bash, zsh, fish, and powershell
- Hook mechanism with `pre_load`, `post_load`, `pre_clear`, `post_clear` support
- Environment variable passing to hooks (`GS_SLOT_NAME`, `GS_SLOT_PATH`, `GS_BRANCH`, `GS_REPO_ROOT`, `GS_ACTION`)
- Hook timeout support

[0.2.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.2.0
[0.1.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.1.1
[0.1.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.1.0
