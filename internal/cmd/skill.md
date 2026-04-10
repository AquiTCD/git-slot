---
name: git-slot-workflow
description: Guidelines for managing parallel development environments using the git-slot worktree wrapper.
---

# Git-Slot: Parallel Worktree Management

**Context**:
`git-slot` is a CLI tool wrapping `git worktree` to manage fixed-path, persistent workspaces called "slots" (e.g., `main-work`, `hotfix`). It prevents IDE index thrashing by keeping paths stable.

**Trigger**:
Apply these rules when the user asks to use a slot, work in a named slot, or do parallel branch work with `git-slot` (e.g. "work in slot `main-work`", "use the hotfix slot").

## 1. Core Constraints
1. **Conditional Root Ban**: When working within a slot task, do NOT execute `git checkout` or `git switch` in the repository root. File modifications for that task must be restricted to the slot's directory.
2. **No Interactive TUI/Shells**: To prevent agent freezes from standard input blocking, NEVER execute interactive commands like `gsl` (bare), `git slot` (bare), or `git slot shell`.
3. **Sub-shell Prevention**: You MUST append `--no-shell` to all `git slot set` commands to safely bypass the user's potential `launch_shell=true` global configuration.
4. **No slot shell for agents**: Do not use `git slot shell` or rely on `launch_shell` for agent-driven flows (TTY / stdin issues).
5. **Terminal commands MUST go through `git slot exec`**: For any slot-scoped task, every build/test/lint/package-manager/script invocation you run in a **terminal** MUST be executed as **`git slot exec …`** (see §2.1). This applies whether or not `[[slots.env]]` / `GSL_*` look "necessary" — using `exec` is the default and is not harmful when env extras are empty; it keeps **cwd**, **GSL_***, and **`[[slots.env]]`** consistent. **Do not** run bare commands inside the slot tree after only `cd` (e.g. `npm test`, `go test`, `make`) as a way to skip `exec`. The only exceptions are: the user explicitly tells you to run without `exec`, or a command truly cannot run under `exec` (report that to the user).

## 2. Slot Lifecycle Workflow

Because you are using `--no-shell`, `git-slot` will **not** automatically change your directory or start a subshell. You must handle navigation manually.

### 2.1 `git slot exec` and `git slot which` (non-interactive)

**Required** way to run terminal commands with the correct slot **cwd** and **environment** (`GSL_*` plus `[[slots.env]]`) without `GSL_SHELL_SESSION` and without a TTY-consuming subshell.

- **`git slot exec <slot> -- <command> [args...]`** — Run inside that slot's worktree with merged env. **Preferred when your shell cwd is not yet that worktree** (e.g. repo root). Example: `git slot exec main-work -- npm test`
- **`git slot exec -- <command> [args...]`** — Slot inferred from `cwd` (same rule as `which`). Use after you have `cd`'d into the slot worktree **only** together with the rule above: **still wrap** terminal commands; do not treat `cd` as permission to run raw `npm` / `go` / `make`.
- **`git slot which`** — Prints the configured slot name for the current git worktree root (one line). Use to confirm you are in a slot tree; exits non-zero if `cwd` is not a slot worktree.

**Hard rule**: `--` is required before the command (e.g. `exec -- npm run build`, not `exec npm run build`).

### Step 1: Mount the Branch
If the user names a slot but not a branch, check whether the slot is already active (e.g. `git slot status`) before mounting.
- For an existing branch:
  `git slot set <slot_name> <branch_name> --no-shell`
- To create and mount a new branch:
  `git slot set <slot_name> -c <new_branch_name> --no-shell`

### Step 2: Resolve Path & Navigate (for editor / context)
`cd` into the slot worktree when you need the agent's **file operations** to target that tree (paths, reads, writes). Get the path with `git slot set <slot_name>` (prints absolute path), then `cd` there. **This step does not replace `git slot exec` for terminal commands** — see constraint 5 and §2.1.

### Step 3: Execute Task
Perform file edits in the slot directory as needed. **Every terminal-driven** build, test, install, linter, or script **MUST** use `git slot exec <slot> -- …` or, if your cwd is already that slot's root, `git slot exec -- …`. Do not use "we already `cd`'d" as a reason to omit `exec`.

### Step 4: End of Task 
When your coding/review task is completed:
- **Do not automatically clear the slot** unless explicitly instructed by the user.
- Leave the workspace as-is and report back to the user that the task is done. If the user asks you to clean up, run `git slot clear <slot_name>`. Ask for permission before using `-f` if the slot is dirty.

## 3. References
If you need to perform advanced operations (like checking out a PR into a slot via the `--pr` flag), refer to the repository's `README.md` or `docs/specs/` for detailed command references.
