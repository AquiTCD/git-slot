package cmd

import (
	"fmt"

	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/spf13/cobra"
)

var whichCmd = &cobra.Command{
	Use:   "which",
	Short: "Print the slot name for the current git worktree",
	Long: `Prints the git-slot slot name whose worktree root matches the current directory's
git top-level (git rev-parse --show-toplevel). Exits with a non-zero status if the
current directory is not inside a configured slot worktree.`,
	Args: cobra.NoArgs,
	RunE: runWhich,
}

func init() {
	rootCmd.AddCommand(whichCmd)
}

func (a *app) slotNameFromGitCWD() (string, error) {
	d := git.NewExecDetector("")
	top, err := d.RepoRoot()
	if err != nil {
		return "", err
	}
	return slot.SlotNameForWorktreeRoot(a.cfg, a.basePath, a.repoName, top)
}

func runWhich(cmd *cobra.Command, _ []string) error {
	a, err := bootstrap()
	if err != nil {
		return err
	}
	name, err := a.slotNameFromGitCWD()
	if err != nil {
		return exitIfNotSlotWorktree(err, "not inside a configured git-slot worktree")
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), name)
	return err
}
