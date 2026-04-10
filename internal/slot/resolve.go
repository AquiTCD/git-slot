package slot

import (
	"errors"
	"path/filepath"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/pathutil"
)

// ErrNotASlotWorktree means the given path is not the root of a configured git-slot worktree.
var ErrNotASlotWorktree = errors.New("not inside a configured git-slot worktree")

// SlotNameForWorktreeRoot returns the slot name whose worktree root equals worktreeTop
// (typically from git rev-parse --show-toplevel), or ErrNotASlotWorktree.
func SlotNameForWorktreeRoot(cfg *config.Config, basePath, repoName, worktreeTop string) (string, error) {
	if cfg == nil || len(cfg.Slots) == 0 {
		return "", ErrNotASlotWorktree
	}
	top := filepath.Clean(worktreeTop)
	for _, def := range cfg.Slots {
		if def.Name == "" {
			continue
		}
		p := pathutil.ResolveSlotPath(basePath, repoName, def.Name)
		if filepath.Clean(p) == top {
			return def.Name, nil
		}
	}
	return "", ErrNotASlotWorktree
}
