---
name: git-slot-workflow
description: Guidelines for managing parallel development environments using the git-slot worktree wrapper.
---

# Git-Slot: Parallel Worktree Management

**Context**:
`git-slot` is a CLI tool wrapping `git worktree` to manage fixed-path, persistent workspaces called "slots" (e.g., `main-work`, `hotfix`). It prevents IDE index thrashing by keeping paths stable.

**Trigger**:
Apply these rules ONLY when you are instructed to "use a slot", "work in slot X", or when parallel branch development via `git-slot` is explicitly requested.

## 1. Core Constraints
1. **Conditional Root Ban**: When working within a slot task, do NOT execute `git checkout` or `git switch` in the repository root. File modifications for that task must be restricted to the slot's directory.
2. **No Interactive TUI/Shells**: To prevent agent freezes from standard input blocking, NEVER execute interactive commands like `gsl` (bare), `git slot` (bare), or `git slot shell`.
3. **Sub-shell Prevention**: You MUST append `--no-shell` to all `git slot set` commands to safely bypass the user's potential `launch_shell=true` global configuration.
4. **Slot env without shells**: Do not use `git slot shell` or rely on `launch_shell` for agents (TTY / stdin issues). When `[[slots.env]]` or `GSL_*` must apply, **MUST** use **`git slot exec`** (§2.1), not a plain subprocess with only `cd`.
5. **Default command wrapper in-slot**: Once your `cwd` is inside the **correct** slot worktree for the task (and it matches the slot you are working in), **SHOULD** run build/test/package-manager and other CLI invocations through **`git slot exec -- …`** so working directory and merged env stay aligned with git-slot. Deviate only when the user explicitly asks otherwise or a command cannot run that way.

## 2. Slot Lifecycle Workflow

Because you are using `--no-shell`, `git-slot` will **not** automatically change your directory or start a subshell. You must handle navigation manually.

### 2.1 `git slot exec` and `git slot which` (non-interactive)

Primary way to run commands with the right slot **cwd** and **environment** (`GSL_*` plus `[[slots.env]]`) without `GSL_SHELL_SESSION` and without a TTY-consuming subshell.

- **`git slot exec <slot> -- <command> [args...]`** — Run inside that slot's worktree with merged env. Use from anywhere when you have not `cd`'d yet. Example: `git slot exec main-work -- npm test`
- **`git slot exec -- <command> [args...]`** — Slot inferred from `cwd` (same rule as `which`). **This is the usual form after Step 2** when you are already in the slot tree: default wrapper for CLI work (see constraint 5).
- **`git slot which`** — Prints the configured slot name for the current git worktree root (one line). Use to confirm you are in a slot tree or to pass the name to scripts; exits non-zero if `cwd` is not a slot worktree.

**Hard rule**: `--` is required before the command (e.g. `exec -- npm run build`, not `exec npm run build`).

### Step 1: Mount the Branch
- For an existing branch:
  `git slot set <slot_name> <branch_name> --no-shell`
- To create and mount a new branch:
  `git slot set <slot_name> -c <new_branch_name> --no-shell`

### Step 2: Resolve Path & Navigate (Mandatory)
You MUST manually navigate into the slot to perform work.
- Get the path: Execute `git slot set <slot_name>` to print the absolute path.
- Navigate: `cd` into that absolute path.

### Step 3: Execute Task
Perform all file edits, script executions, and tests strictly within this slot's directory. **Run CLI commands via `git slot exec -- …` by default** (constraint 5). File edits in the editor do not need `exec`; terminal-driven builds/tests/installs do.

### Step 4: End of Task 
When your coding/review task is completed:
- **Do not automatically clear the slot** unless explicitly instructed by the user.
- Leave the workspace as-is and report back to the user that the task is done. If the user asks you to clean up, run `git slot clear <slot_name>`. Ask for permission before using `-f` if the slot is dirty.

## 3. References
If you need to perform advanced operations (like checking out a PR into a slot via the `--pr` flag), refer to the repository's `README.md` or `docs/specs/` for detailed command references.
