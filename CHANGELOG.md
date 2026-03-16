# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.5.0] - 2026-03-16

### Changed

- **Breaking**: Migrated from flag-based to subcommand-based CLI. All operations are now subcommands:
  - `git slot set <slot> [branch]` (was: `git slot <slot> [branch]`)
  - `git slot list` (was: `git slot --list`)
  - `git slot clear <slot>` (was: `git slot -d <slot>`)
  - `git slot swap <A> <B>` (was: `git slot --swap <A> <B>`)
  - `git slot status [slot]` (was: `git slot --status [slot]`)
  - `git slot init` (was: `git slot --init`)
  - `git slot hook` (was: `git slot --hook`)
  - `git slot root` (was: `git slot --eject`)
- Added `-v` shorthand for `--version` on root command.
- Added `-f` shorthand for `--force` on `set`, `clear`, and `init` subcommands.
- gsl wrapper usage updated: `gsl set <slot>`, `gsl root` (was: `gsl <slot>`, `gsl -e`).

### Fixed

- `--eject` / `root` now correctly returns the main repo root even from linked worktrees.

## [0.4.1] - 2026-03-16

### Added

- Unit tests for errutil, git/ignored, hook runner (link/copy), and tui/hook_helper.
- Mock-based integration tests for cmd (runList, runMount, runClear, hook helpers, resolveConfigPath).

### Changed

- slot: Extracted SlotManager interface; renamed Load→Mount, LoadOptions→MountOptions; added populateSlot/worktreeMap and enrichStatus helpers to reduce duplication.
- cmd: Pass global/force flags explicitly; accept SlotManager interface; extracted resolveConfigPath, buildPostMountHooks, generateHooksTOML from updateConfigWithHooks; renamed cmd_load→cmd_mount, runLoad→runMount.
- hook: Extracted resolveAndPrepare to deduplicate link/copy setup.

## [0.4.0] - 2026-03-16

### Added

- Powerful hook management via `git slot --hook`:
  - Automatically discovers ignored/untracked files in your repository.
  - Interactive TUI to choose between `Link`, `Copy`, or `None` for each file.
  - Performs surgical TOML updates to preserve user-added comments and other configurations.
- Intelligent configuration merging:
  - Slots are now merged by name (project-level slots overwrite or extend global slots).
  - Hook configurations are merged field-by-field, allowing project hooks to coexist with global ones.
- Enhanced hook execution capabilities:
  - Directory support: Correctly handles symbolic links and copies for entire directories.
  - Improved error reporting: Hook failures now include the command's exit code for easier debugging.

### Changed

- Updated terminology: Changed `load` to `mount` throughout the codebase, including configuration fields (`pre_mount`, `post_mount`).
- Hook structure upgrade: Hook actions are now slices (`[]HookAction`), allowing multiple actions (link, copy, run) for the same hook trigger.
- Configuration field rename: `gwq_basedir` is now `slots_base_path` for better clarity.

### Fixed

- Configuration duplication: Repeatedly running `git slot --hook` no longer results in duplicate configuration blocks.
- .gitignore patterns: Correctly handles directory symbolic links in Git.

## [0.3.0] - 2026-03-15

### Added

- Interactive slot filter is now enabled by default.
  - Removed optional `[tui].filter` config in favor of an always-on experience.
  - Optimized TUI navigation: `ctrl+j/k` always available for selection while typing.
- Standardized error handling and exit codes via new `errutil` package.
  - Config errors now return exit code 2.
  - Git repository detection errors return exit code 3.

### Changed

- Major architectural refactor: Split monolithic CLI logic into separate command files (`cmd_list.go`, `cmd_load.go`, etc.) for better maintainability.
- Performance optimization: Introduced caching for `git worktree list` calls in `slot.Manager`, reducing redundant Git executions.
- Environment variable consistency: Renamed internal hook variables to use `GSL_` prefix (e.g., `GSL_SLOT_NAME`), aligning with the recommended `gsl` shorthand.

### Fixed

- Improved color detection logic: Correctly respects `CLICOLOR_FORCE` and forces color output even when terminal detection is ambiguous (e.g., inside shell wrappers).


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

[0.5.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.5.0
[0.4.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.4.1
[0.4.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.4.0
[0.3.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.3.0
[0.2.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.2.0
[0.1.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.1.1
[0.1.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.1.0
