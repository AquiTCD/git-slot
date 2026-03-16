package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/hook"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/AquiTCD/git-slot/internal/slot"
)

func runClear(a *app, slotName string, force bool, out io.Writer) error {
	hookRunner := hook.NewRunner(out, os.Stderr)
	slotPath := pathutil.ResolveSlotPath(a.basePath, slotName)

	env := hook.HookEnv{
		SlotName: slotName,
		SlotPath: slotPath,
		RepoRoot: a.repoRoot,
		Action:   "clear",
	}

	if err := hookRunner.Run(a.cfg.Hooks.PreClear, env); err != nil {
		return fmt.Errorf("pre-clear hook: %w", err)
	}

	if err := a.mgr.Clear(slotName, slot.ClearOptions{Force: force}); err != nil {
		return err
	}

	if err := hookRunner.Run(a.cfg.Hooks.PostClear, env); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: post-clear hook: %v\n", err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Slot '%s' is now empty.\n", slotName)
	return nil
}
