package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/gslenv"
	"github.com/AquiTCD/git-slot/internal/hook"
	tea "github.com/charmbracelet/bubbletea"
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

// checkExecAllowedInSlotShell enforces that from an interactive slot shell, exec may only
// target the current slot (by name or by cwd resolution).
func checkExecAllowedInSlotShell(explicitSlot bool, slotArg, resolvedSlotName string) error {
	if !isInsideSlotShell() {
		return nil
	}
	current := os.Getenv("GSL_SLOT_NAME")
	if explicitSlot {
		if slotArg != current {
			return fmt.Errorf("%w: from a slot shell, only current slot %q is allowed (got %q)", ErrShellNested, current, slotArg)
		}
		return nil
	}
	if resolvedSlotName != current {
		return fmt.Errorf("%w: cwd resolves to slot %q but this shell session is for slot %q", ErrShellNested, resolvedSlotName, current)
	}
	return nil
}

func wantShell(cfg *config.Config, noShell bool) bool {
	if noShell || gslenv.FromWrapper() {
		return false
	}
	return cfg.LaunchShell != nil && *cfg.LaunchShell
}

func newHookContext(a *app, slotName, branch, action string, out io.Writer) (*hook.Runner, hook.HookEnv) {
	runner := hook.NewRunner(out, os.Stderr)
	slotPath := a.mgr.SlotPath(slotName)
	env := buildHookEnv(a.cfg, slotName, slotPath, branch, a.repoRoot, action)
	return runner, env
}

type abortable interface {
	Aborted() bool
}

func runTUI[M abortable](model tea.Model) (M, bool, error) {
	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		var zero M
		return zero, false, fmt.Errorf("interactive mode: %w", err)
	}
	m, ok := finalModel.(M)
	if !ok {
		var zero M
		return zero, false, fmt.Errorf("internal error: unexpected model type %T", finalModel)
	}
	if m.Aborted() {
		return m, true, nil
	}
	return m, false, nil
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
