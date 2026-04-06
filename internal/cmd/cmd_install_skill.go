package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillRulesMD string

var installSkillCmd = &cobra.Command{
	Use:   "install-skill [output-dir]",
	Short: "Output AI Agent guidelines (skill) for git-slot",
	Long: `Outputs a markdown file containing guidelines for AI coding agents 
(like Claude Code, Cursor, Windsurf) to properly use git-slot arrays and avoid common pitfalls.

By default, without arguments, it outputs to .agents/skills/git-slot-workflow/SKILL.md
if the --stdout or --append flags are not used.`,
	RunE: runInstallSkill,
}

func init() {
	installSkillCmd.Flags().Bool("stdout", false, "Output directly to standard output")
	installSkillCmd.Flags().String("append", "", "Append the rules to an existing file (e.g. .cursorrules)")
	rootCmd.AddCommand(installSkillCmd)
}

func runInstallSkill(cmd *cobra.Command, args []string) error {
	printStdout, _ := cmd.Flags().GetBool("stdout")
	appendFile, _ := cmd.Flags().GetString("append")

	content := skillRulesMD

	// Case 1: Print to stdout
	if printStdout {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), content)
		return nil
	}

	// Case 2: Append to a file
	if appendFile != "" {
		f, err := os.OpenFile(appendFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file for appending: %w", err)
		}
		defer func() { _ = f.Close() }()

		if _, err := f.WriteString("\n\n" + content); err != nil {
			return fmt.Errorf("failed to append content: %w", err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully appended git-slot skill to %s\n", appendFile)
		return nil
	}

	// Case 3: Write to output directory
	outDir := ".agents/skills/git-slot-workflow"
	if len(args) > 0 {
		outDir = args[0]
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outDir, err)
	}

	targetFile := filepath.Join(outDir, "SKILL.md")

	// Create or overwrite
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully installed git-slot skill to %s\n", targetFile)
	return nil
}
