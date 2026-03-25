package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear <slot>",
	Short: "Clear (remove) a slot's worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		a, err := bootstrap()
		if err != nil {
			return err
		}
		return runClear(a, args[0], force, cmd.OutOrStdout())
	},
}

func init() {
	clearCmd.Flags().BoolP("force", "f", false, "Skip confirmation for destructive actions")

	rootCmd.AddCommand(clearCmd)
}

func runClear(a *app, slotName string, force bool, out io.Writer) error {
	hookRunner, env := newHookContext(a, slotName, "", "clear", out)

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
