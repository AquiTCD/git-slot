package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
)

type mountOptions struct {
	createBranch bool
	force        bool
	noShell      bool
}

func runMount(a *app, slotName, branchName string, opts mountOptions, out io.Writer) error {
	shouldLaunchShell := wantShell(a.cfg, opts.noShell)

	if shouldLaunchShell {
		if err := checkShellNestingForSet(slotName); err != nil {
			return err
		}
		if isInsideSlotShell() && os.Getenv("GSL_SLOT_NAME") == slotName {
			shouldLaunchShell = false
		}
	}

	hookRunner, env := newHookContext(a, slotName, branchName, "mount", out)

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

	return runGetPath(a.mgr, slotName, out)
}

func runGetPath(mgr slot.SlotManager, slotName string, out io.Writer) error {
	path, err := mgr.GetPath(slotName)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, path)
	return nil
}

// runGetPathOrLaunchShell prints the worktree path, or launches a slot shell when
// launch_shell is set and the slot already has a branch mounted. This unifies
// `git slot set <slot>` (one argument), TUI selection of an active slot only, and
// runMount's "same slot inside shell → print path" behaviour.
func runGetPathOrLaunchShell(a *app, slotName string, noShell bool, out io.Writer) error {
	if wantShell(a.cfg, noShell) {
		if err := checkShellNestingForSet(slotName); err != nil {
			return err
		}
		if isInsideSlotShell() && os.Getenv("GSL_SLOT_NAME") == slotName {
			return runGetPath(a.mgr, slotName, out)
		}
		st, err := a.mgr.Status(slotName)
		if err != nil {
			return err
		}
		if st.State != slot.SlotEmpty {
			return launchSlotShell(a, slotName)
		}
	}
	return runGetPath(a.mgr, slotName, out)
}

func runInteractive(a *app, force, noShell bool, out io.Writer) error {
	if wantShell(a.cfg, noShell) {
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

	m, aborted, err := runTUI[tui.Model](model)
	if err != nil {
		return err
	}
	if aborted {
		return nil
	}

	result, ok := m.GetResult()
	if !ok {
		return nil
	}

	return handleInteractiveResult(a, result, force, noShell, out)
}

// handleInteractiveResult applies TUI selection. When the chosen slot is already
// active, BranchName is empty (see tui.Model.selectCurrentSlot); launch_shell must
// still open a slot shell in that case, not only print the path.
func handleInteractiveResult(a *app, result tui.Result, force, noShell bool, out io.Writer) error {
	if result.BranchName == "" {
		return runGetPathOrLaunchShell(a, result.SlotName, noShell, out)
	}

	return runMount(a, result.SlotName, result.BranchName, mountOptions{
		createBranch: result.CreateBranch,
		force:        force,
		noShell:      noShell,
	}, out)
}
