package cmd

import (
	"fmt"
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a template config file",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		global, _ := cmd.Flags().GetBool("global")
		force, _ := cmd.Flags().GetBool("force")
		return runInit(global, force)
	},
}

func init() {
	initCmd.Flags().BoolP("global", "g", false, "Generate global config instead of project config")
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing config file")

	rootCmd.AddCommand(initCmd)
}

func runInit(global, force bool) error {
	detector := git.NewExecDetector("")
	repoRoot, _ := detector.RepoRoot()

	path, err := config.Init(config.InitOptions{
		Global: global,
		Force:  force,
	}, repoRoot)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stderr, "Created %s with template configuration.\n", path)
	return nil
}
