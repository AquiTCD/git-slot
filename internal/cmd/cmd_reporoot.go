package cmd

import (
	"fmt"

	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/spf13/cobra"
)

var repoRootCmd = &cobra.Command{
	Use:   "root",
	Short: "Print the repository root path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		detector := git.NewExecDetector("")
		mainRoot, err := detector.MainRepoRoot()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), mainRoot)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(repoRootCmd)
}
