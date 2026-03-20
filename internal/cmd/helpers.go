package cmd

import (
	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/hook"
)

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
