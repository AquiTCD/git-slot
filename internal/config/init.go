package config

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrConfigExists = errors.New("configuration file already exists")

const configTemplate = `# git-slot.toml — Git Slot configuration
# See: https://github.com/AquiTCD/git-slot

# gwq base directory (same as gwq's worktree.basedir)
# Default: ~/worktrees
# gwq_basedir = "~/worktrees"

# Define your slots below.
# Add as many [[slots]] entries as you need.

[[slots]]
name = "slot-1"
# icon = "🔧"

[[slots]]
name = "slot-2"
# icon = "🔥"

[[slots]]
name = "slot-3"
# icon = "💧"

# Optional: hooks
# [hooks]
# post_load = ".git-slot/hooks/post-load.sh"
# post_clear = ".git-slot/hooks/post-clear.sh"
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
