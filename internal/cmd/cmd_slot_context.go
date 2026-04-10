package cmd

import (
	"errors"
	"fmt"

	"github.com/AquiTCD/git-slot/internal/errutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/slotenv"
)

// exitIfNotSlotWorktree maps slot.ErrNotASlotWorktree to a non-zero exit error with msg.
func exitIfNotSlotWorktree(err error, msg string) error {
	if errors.Is(err, slot.ErrNotASlotWorktree) {
		return errutil.NewExitError(msg, 1)
	}
	return err
}

// loadActiveSlotContext returns status, SlotInfo, and user env map for a mounted slot.
// It errors if the slot is empty or status lookup fails.
func loadActiveSlotContext(a *app, slotName string) (*slot.SlotStatus, slotenv.SlotInfo, map[string]string, error) {
	st, err := a.mgr.Status(slotName)
	if err != nil {
		return nil, slotenv.SlotInfo{}, nil, err
	}
	if st.State == slot.SlotEmpty {
		return nil, slotenv.SlotInfo{}, nil, fmt.Errorf("slot '%s' is empty; mount a branch first with 'git slot set'", slotName)
	}
	var userEnv map[string]string
	if def := a.cfg.FindSlot(slotName); def != nil {
		userEnv = def.Env
	}
	info := slotenv.SlotInfo{
		SlotName: slotName,
		SlotPath: st.Path,
		Branch:   st.Branch,
		RepoRoot: a.repoRoot,
	}
	return st, info, userEnv, nil
}
