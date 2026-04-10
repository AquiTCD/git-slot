// Package slotenv builds and merges GSL_* environment variable sets for slot shells and exec.
package slotenv

import "strings"

type SlotInfo struct {
	SlotName string
	SlotPath string
	Branch   string
	RepoRoot string
}

// BuildSlotExecEnv returns GSL_* and user slot env for one-shot commands (git slot exec).
// It does not set GSL_SHELL_SESSION so child processes are not treated as an interactive slot shell.
func BuildSlotExecEnv(info SlotInfo, userEnv map[string]string) []string {
	base := []string{
		"GSL_SLOT_NAME=" + info.SlotName,
		"GSL_SLOT_PATH=" + info.SlotPath,
		"GSL_BRANCH=" + info.Branch,
		"GSL_REPO_ROOT=" + info.RepoRoot,
	}
	return mergeUserSlotEnv(base, userEnv)
}

// BuildSlotEnv returns env for an interactive slot shell, including GSL_SHELL_SESSION=1.
func BuildSlotEnv(info SlotInfo, userEnv map[string]string) []string {
	base := []string{
		"GSL_SLOT_NAME=" + info.SlotName,
		"GSL_SLOT_PATH=" + info.SlotPath,
		"GSL_BRANCH=" + info.Branch,
		"GSL_REPO_ROOT=" + info.RepoRoot,
		"GSL_SHELL_SESSION=1",
	}
	return mergeUserSlotEnv(base, userEnv)
}

func mergeUserSlotEnv(gslVars []string, userEnv map[string]string) []string {
	gslKeys := make(map[string]struct{}, len(gslVars))
	for _, v := range gslVars {
		key, _, _ := strings.Cut(v, "=")
		gslKeys[key] = struct{}{}
	}

	for k, v := range userEnv {
		if _, reserved := gslKeys[k]; reserved {
			continue
		}
		gslVars = append(gslVars, k+"="+v)
	}

	return gslVars
}

func MergeEnvWithOS(osEnv, slotEnv []string) []string {
	overrideKeys := make(map[string]string, len(slotEnv))
	for _, entry := range slotEnv {
		key, _, _ := strings.Cut(entry, "=")
		overrideKeys[key] = entry
	}

	merged := make([]string, 0, len(osEnv)+len(slotEnv))
	for _, entry := range osEnv {
		key, _, _ := strings.Cut(entry, "=")
		if _, overridden := overrideKeys[key]; overridden {
			continue
		}
		merged = append(merged, entry)
	}

	merged = append(merged, slotEnv...)
	return merged
}
