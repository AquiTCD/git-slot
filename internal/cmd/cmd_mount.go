package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/hook"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

type mountOptions struct {
	createBranch bool
	force        bool
	noShell      bool
}

func runMount(a *app, slotName, branchName string, opts mountOptions, out io.Writer) error {
	shouldLaunchShell := a.cfg.LaunchShell != nil && *a.cfg.LaunchShell && !opts.noShell

	if shouldLaunchShell {
		if err := checkShellNestingForSet(slotName); err != nil {
			return err
		}
		if isInsideSlotShell() && os.Getenv("GSL_SLOT_NAME") == slotName {
			shouldLaunchShell = false
		}
	}

	hookRunner := hook.NewRunner(out, os.Stderr)
	slotPath := pathutil.ResolveSlotPath(a.basePath, slotName)
	env := buildHookEnv(a.cfg, slotName, slotPath, branchName, a.repoRoot, "mount")

	if err := hookRunner.Run(a.cfg.Hooks.PreMount, env); err != nil {
		return fmt.Errorf("pre-mount hook: %w", err)
	}

	if err := a.mgr.Mount(slotName, branchName, slot.MountOptions{
		CreateBranch: opts.createBranch,
		Force:        opts.force,
	}); err != nil {
		return err
	}

	if err := hookRunner.Run(a.cfg.Hooks.PostMount, env); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: post-mount hook: %v\n", err)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Slot '%s' is ready.\n", slotName)

	if shouldLaunchShell {
		return launchSlotShell(a, slotName)
	}

	path, err := a.mgr.GetPath(slotName)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, path)
	return nil
}

func isInsideSlotShell() bool {
	return os.Getenv("GSL_SHELL_SESSION") != ""
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

func runGetPath(mgr slot.SlotManager, slotName string, out io.Writer) error {
	path, err := mgr.GetPath(slotName)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, path)
	return nil
}

func runInteractive(a *app, force, noShell bool, out io.Writer) error {
	if a.cfg.LaunchShell != nil && *a.cfg.LaunchShell && !noShell {
		if err := checkShellNesting(); err != nil {
			return err
		}
	}

	slots, err := a.mgr.List()
	if err != nil {
		return err
	}
	if len(slots) == 0 {
		_, _ = fmt.Fprintln(out, "No slots defined.")
		return nil
	}

	branches, err := a.wt.ListBranches()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: could not list branches: %v\n", err)
		branches = nil
	}

	noColor := tui.IsNoColor()
	model := tui.NewInteractiveModel(slots, branches, noColor)

	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("interactive mode: %w", err)
	}

	m, ok := finalModel.(tui.Model)
	if !ok {
		return fmt.Errorf("internal error: unexpected model type %T", finalModel)
	}
	if m.Aborted() {
		return nil
	}

	result, ok := m.GetResult()
	if !ok {
		return nil
	}

	if result.BranchName == "" {
		return runGetPath(a.mgr, result.SlotName, out)
	}

	return runMount(a, result.SlotName, result.BranchName, mountOptions{
		createBranch: result.CreateBranch,
		force:        force,
		noShell:      noShell,
	}, out)
}
