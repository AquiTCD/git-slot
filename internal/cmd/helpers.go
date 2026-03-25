package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/hook"
	"github.com/AquiTCD/git-slot/internal/pathutil"
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

func wantShell(cfg *config.Config, noShell bool) bool {
	return cfg.LaunchShell != nil && *cfg.LaunchShell && !noShell
}

func newHookContext(a *app, slotName, branch, action string, out io.Writer) (*hook.Runner, hook.HookEnv) {
	runner := hook.NewRunner(out, os.Stderr)
	slotPath := pathutil.ResolveSlotPath(a.basePath, slotName)
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
