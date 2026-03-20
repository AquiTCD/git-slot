package config

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrConfigExists = errors.New("configuration file already exists")

const configTemplate = `# git-slot.toml — Git Slot configuration
# See: https://github.com/AquiTCD/git-slot

# Base directory for slots.
# Default: ~/worktrees
# slots_base_path = "~/worktrees"

# Launch a sub-shell after mounting a slot (default: false).
# When true, 'set' and interactive TUI will open a sub-shell
# with GSL_* environment variables inside the slot directory.
# launch_shell = true

# Define your slots below.
# Add as many [[slots]] entries as you need.

[[slots]]
name = "slot-1"
# icon = "🔧"
# [slots.env]
# PORT = "3001"

[[slots]]
name = "slot-2"
# icon = "🔥"
# [slots.env]
# PORT = "3002"

# Optional: hooks
# [hooks]
# Generic post-mount hooks (links, copies, commands)
# [[hooks.post_mount]]
# type = "link"
# source = "$GSL_REPO_ROOT/.env"
# dest = "$GSL_SLOT_PATH/.env"

# [[hooks.post_mount]]
# type = "run"
# command = "npm install"
`

type InitOptions struct {
	Global bool
	Force  bool
}

func Init(opts InitOptions, repoRoot string) (string, error) {
	var path string
	if opts.Global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, ".config", "git-slot", "config.toml")
	} else {
		path = filepath.Join(repoRoot, "git-slot.toml")
	}

	if !opts.Force {
		if _, err := os.Stat(path); err == nil {
			return "", ErrConfigExists
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(configTemplate), 0o644); err != nil {
		return "", err
	}

	return path, nil
}
