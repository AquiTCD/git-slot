package cmd

import (
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/hook"
)

func isInsideSlotShell() bool {
	return os.Getenv("GSL_SHELL_SESSION") != ""
}

func checkShellNesting() error {
	if isInsideSlotShell() {
		return ErrShellNested
	}
	return nil
}

func checkShellNestingForSet(targetSlot string) error {
	if !isInsideSlotShell() {
		return nil
	}
	currentSlot := os.Getenv("GSL_SLOT_NAME")
	if currentSlot == targetSlot {
		return nil
	}
	return ErrShellNested
}

func buildHookEnv(cfg *config.Config, slotName, slotPath, branch, repoRoot, action string) hook.HookEnv {
	var userEnv map[string]string
	if def := cfg.FindSlot(slotName); def != nil {
		userEnv = def.Env
	}
	return hook.HookEnv{
		SlotName: slotName,
		SlotPath: slotPath,
		Branch:   branch,
		RepoRoot: repoRoot,
		Action:   action,
		UserEnv:  userEnv,
	}
}
