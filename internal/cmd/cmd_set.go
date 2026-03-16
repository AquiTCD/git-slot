package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <slot> [branch]",
	Short: "Set a branch into a slot, or print the slot path",
	Long: `Set a branch into a named slot, or print the slot's worktree path.

With one argument, prints the slot's worktree path.
With two arguments, mounts the specified branch into the slot.
Use -c/--create to create a new branch and mount it.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runSet,
}

func init() {
	setCmd.Flags().StringP("create", "c", "", "Create a new branch and mount into slot")
	setCmd.Flags().StringP("branch", "b", "", "Alias for --create")
	setCmd.Flags().BoolP("force", "f", false, "Skip confirmation for destructive actions")

	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	create, _ := cmd.Flags().GetString("create")
	branch, _ := cmd.Flags().GetString("branch")
	force, _ := cmd.Flags().GetBool("force")

	if branch != "" {
		if create != "" && create != branch {
			return fmt.Errorf("--create and --branch cannot specify different values")
		}
		create = branch
	}

	a, err := bootstrap()
	if err != nil {
		return err
	}

	slotName := args[0]
	out := cmd.OutOrStdout()

	if len(args) == 2 {
		return runMount(a, slotName, args[1], false, force, out)
	}
	if create != "" {
		return runMount(a, slotName, create, true, force, out)
	}
	return runGetPath(a.mgr, slotName, out)
}
