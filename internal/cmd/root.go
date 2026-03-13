package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagList    bool
	flagClear   string
	flagSwap    []string
	flagStatus  string
	flagInit    bool
	flagGlobal  bool
	flagCreate  string
	flagBranch  string
	flagForce   bool
	flagJSON    bool
	flagVersion bool
)

var rootCmd = &cobra.Command{
	Use:   "git-slot [slot] [branch]",
	Short: "Manage git worktrees as fixed slots",
	Long: `git-slot manages git worktrees as fixed, named slots defined in TOML configuration.
Load branches into slots, clear them, swap between them, and more.

Usage as a git subcommand:
  git slot <slot> <branch>       Load an existing branch into a slot
  git slot <slot> -c <branch>    Create a new branch and load it into a slot
  git slot <slot>                Print the slot's worktree path

Management flags:
  git slot -l, --list            List all slots and their status
  git slot -d, --clear <slot>    Clear (remove) a slot's worktree
  git slot -s, --swap <A> <B>    Swap branches between two slots
  git slot --status [slot]       Show detailed slot status
  git slot --init                Generate a template config file`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          run,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagList, "list", "l", false, "List all slots and their status")
	rootCmd.Flags().StringVarP(&flagClear, "clear", "d", "", "Clear (remove) a slot's worktree")
	rootCmd.Flags().StringSliceVarP(&flagSwap, "swap", "s", nil, "Swap branches between two slots")
	rootCmd.Flags().StringVar(&flagStatus, "status", "", "Show detailed slot status")
	rootCmd.Flags().BoolVar(&flagInit, "init", false, "Generate a template config file")
	rootCmd.Flags().BoolVar(&flagGlobal, "global", false, "Used with --init to generate global config")
	rootCmd.Flags().StringVarP(&flagCreate, "create", "c", "", "Create a new branch and load into slot")
	rootCmd.Flags().StringVarP(&flagBranch, "branch", "b", "", "Alias for --create")
	rootCmd.Flags().BoolVar(&flagForce, "force", false, "Skip confirmation for destructive actions")
	rootCmd.Flags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&flagVersion, "version", false, "Print version information")

	_ = rootCmd.Flags().MarkHidden("global")
}

func run(cmd *cobra.Command, args []string) error {
	if flagVersion {
		printVersion(cmd.OutOrStdout())
		return nil
	}

	out := cmd.OutOrStdout()

	if flagList {
		_, _ = fmt.Fprintln(out, "not implemented yet: --list")
		return nil
	}

	if flagClear != "" {
		_, _ = fmt.Fprintf(out, "not implemented yet: --clear %s\n", flagClear)
		return nil
	}

	if cmd.Flags().Changed("swap") {
		_, _ = fmt.Fprintf(out, "not implemented yet: --swap %v\n", flagSwap)
		return nil
	}

	if cmd.Flags().Changed("status") {
		_, _ = fmt.Fprintf(out, "not implemented yet: --status %s\n", flagStatus)
		return nil
	}

	if flagInit {
		_, _ = fmt.Fprintln(out, "not implemented yet: --init")
		return nil
	}

	newBranch := flagCreate
	if newBranch == "" {
		newBranch = flagBranch
	}

	switch len(args) {
	case 0:
		return cmd.Help()
	case 1:
		slotName := args[0]
		if newBranch != "" {
			_, _ = fmt.Fprintf(out, "not implemented yet: create branch '%s' in slot '%s'\n", newBranch, slotName)
			return nil
		}
		_, _ = fmt.Fprintf(out, "not implemented yet: get path for slot '%s'\n", slotName)
		return nil
	case 2:
		slotName := args[0]
		branchName := args[1]
		_, _ = fmt.Fprintf(out, "not implemented yet: load branch '%s' into slot '%s'\n", branchName, slotName)
		return nil
	default:
		return fmt.Errorf("too many arguments. Run 'git slot --help' for usage")
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
