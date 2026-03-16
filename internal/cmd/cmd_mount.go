package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/hook"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func runMount(a *app, slotName, branchName string, createBranch, force bool, out io.Writer) error {
	hookRunner := hook.NewRunner(out, os.Stderr)
	slotPath := pathutil.ResolveSlotPath(a.basePath, slotName)

	env := hook.HookEnv{
		SlotName: slotName,
		SlotPath: slotPath,
		Branch:   branchName,
		RepoRoot: a.repoRoot,
		Action:   "mount",
	}

	if err := hookRunner.Run(a.cfg.Hooks.PreMount, env); err != nil {
		return fmt.Errorf("pre-mount hook: %w", err)
	}

	if err := a.mgr.Mount(slotName, branchName, slot.MountOptions{
		CreateBranch: createBranch,
		Force:        force,
	}); err != nil {
		return err
	}

	if err := hookRunner.Run(a.cfg.Hooks.PostMount, env); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: post-mount hook: %v\n", err)
	}

	path, _ := a.mgr.GetPath(slotName)
	_, _ = fmt.Fprintf(os.Stderr, "Slot '%s' is ready.\n", slotName)
	_, _ = fmt.Fprintln(out, path)
	return nil
}

func runGetPath(mgr slot.SlotManager, slotName string, out io.Writer) error {
	path, err := mgr.GetPath(slotName)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, path)
	return nil
}

func runInteractive(a *app, force bool, out io.Writer) error {
	slots, err := a.mgr.List()
	if err != nil {
		return err
	}
	if len(slots) == 0 {
		_, _ = fmt.Fprintln(out, "No slots defined.")
		return nil
	}

	wt := git.NewExecWorktree(a.repoRoot)
	branches, _ := wt.ListBranches()

	noColor := tui.IsNoColor()
	model := tui.NewInteractiveModel(slots, branches, noColor)

	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("interactive mode: %w", err)
	}

	m := finalModel.(tui.Model)
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

	return runMount(a, result.SlotName, result.BranchName, result.CreateBranch, force, out)
}
