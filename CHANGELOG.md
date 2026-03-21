# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.8.1] - 2026-03-21

### Fixed

- `LaunchShell` config field changed from `bool` to `*bool` to fix a one-way-latch bug where `launch_shell = false` in a higher-priority config layer (project or local) could not override `true` set in a lower-priority layer.

## [0.8.0] - 2026-03-21

### Added

- `git slot list` now shows dirty file count (`*N`) and upstream ahead count (`↑N`) for each active slot. Upstream ahead uses local cache only (no fetch). When upstream is not configured, the `↑` indicator is omitted.
- `.git-slot/config.toml` is now supported as a third config layer with the highest priority (global → project → local). Intended to be gitignored and symlinked across worktrees for machine-specific or shared local settings.
- `has_upstream` field added to `--json` output of `git slot list`, disambiguating `ahead_count: 0` from "no upstream configured".

### Changed

- `Slot.IsDirty` bool field removed; dirty state is now derived from `DirtyCount > 0` everywhere, eliminating the possibility of incoherent state.
- `StyleAheadMark` added to TUI styles, distinct from `StyleDirtyMark` (cyan vs red).

## [0.7.2] - 2026-03-20

### Changed

- Renamed config field `slots_base_path` to `wt_base_path` (TOML key change). The field now exclusively uses gwq-compatible semantics: slots are always placed at `{wt_base_path}/{host}/{owner}/{repo}/slots/{slot-name}`. The previous "direct path" behavior (bypassing the host/owner/repo hierarchy) has been removed as it was broken for global config (caused cross-repo slot collisions). Default behavior is unchanged: `~/worktrees/{host}/{owner}/{repo}/slots/{slot-name}`.

## [0.7.1] - 2026-03-20

### Changed

- Comprehensive refactoring across internal packages: extracted `buildHookEnv` helper, defined `HookType` constants, promoted `FindSlot` to a `Config` method, removed dead code in `ResolveSlotsBasePath` and `cmd_shell`.
- Applied Go best practices and linting rules across the codebase (style, lint, additional linters enabled).
- Added GitHub Actions CI test workflow with proper git identity configuration for reproducible test runs.
- Restructured `.agents` directory for Claude Code compatibility.

### Fixed

- Worktree list errors now include stderr context for easier debugging.
- Corrected guard order for dirty/same-branch check in `Mount()` to ensure correct behavior when switching branches.

## [0.7.0] - 2026-03-19

### Added

- **Sub-shell mode**: New `git slot shell [slot]` subcommand launches a sub-shell inside a slot's worktree with GSL_* environment variables and user-defined env automatically exported.
- **`launch_shell` config option**: When `launch_shell = true`, `set` and interactive TUI also launch sub-shells instead of printing paths.
- **Per-slot environment variables**: Define custom key-value pairs per slot via `[slots.env]` in `git-slot.toml`. These are exported to sub-shells and hook commands.
- **`--no-shell` flag**: Suppresses sub-shell launch on `set` when `launch_shell` is enabled (for scripting/pipe use).
- **Nesting detection**: `GSL_SHELL_SESSION` env var prevents shell nesting. Same-slot branch switch is allowed; different-slot switch is blocked with guidance.
- **`slotenv` package**: New internal utility for building and merging slot environment variables.

### Changed

- Hook `run` commands now receive per-slot user env from `[[slots]] env` alongside GSL_* vars.
- `gsl` wrapper delegates `shell` subcommand directly (bypasses `$()` capture for `syscall.Exec` compatibility).
- Init template (`git slot init`) now includes commented examples for `launch_shell` and `[slots.env]`.

## [0.6.0] - 2026-03-16

### Added

- **Tree-based hook TUI**: `git slot hook` now aggregates ignored files by directory, collapsing sibling entries into a single expandable row when 3+ share a parent.
- **fzf-style filter**: Always-on text input filters the hook list in real time, matching by path substring.
- **Tab/Shift+Tab toggle**: Cycles action (None → Link → Copy) on the current item. On aggregated directories, applies to all children.
- **Directory expansion**: Enter/→ drills into directories; children are loaded on demand via `git ls-files`. ←/Esc/Backspace(empty) collapses back.
- **[Mixed] indicator**: When children of a directory have different actions, the parent shows `[Mixed]` instead of a single action.
- `ListIgnoredFilesIn(dir)` method on `ExecWorktree` for on-demand directory expansion.

### Changed

- Hook TUI keybindings unified with interactive slot selector: ctrl+j/k for navigation, ctrl+s to confirm, ctrl+c/esc to cancel.

## [0.5.1] - 2026-03-16

### Changed

- **Breaking**: Mount now uses `git switch` instead of `worktree remove + add` for active slots. Untracked files (node_modules, .env, build caches) are preserved across branch changes within a slot.
- **Breaking**: Removed `swap` command. Use `clear` → `set` to reassign branches between slots.
- `--force` on `set` now maps to `git switch --discard-changes` instead of `worktree remove --force`.
- Same-branch `set` is now a no-op (previously removed and re-added the worktree).

## [0.5.0] - 2026-03-16

### Changed

- **Breaking**: Migrated from flag-based to subcommand-based CLI. All operations are now subcommands:
  - `git slot set <slot> [branch]` (was: `git slot <slot> [branch]`)
  - `git slot list` (was: `git slot --list`)
  - `git slot clear <slot>` (was: `git slot -d <slot>`)
  - `git slot swap <A> <B>` (was: `git slot --swap <A> <B>`) — removed in v0.5.1
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

[0.8.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.8.0
[0.7.2]: https://github.com/AquiTCD/git-slot/releases/tag/v0.7.2
[0.7.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.7.1
[0.8.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.8.1
[0.7.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.7.0
[0.6.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.6.0
[0.5.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.5.1
[0.5.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.5.0
[0.4.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.4.1
[0.4.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.4.0
[0.3.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.3.0
[0.2.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.2.0
[0.1.1]: https://github.com/AquiTCD/git-slot/releases/tag/v0.1.1
[0.1.0]: https://github.com/AquiTCD/git-slot/releases/tag/v0.1.0
