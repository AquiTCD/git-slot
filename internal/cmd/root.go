package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/AquiTCD/git-slot/internal/slot"
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

	if flagInit {
		return runInit(out)
	}

	if len(args) == 0 && !flagList && flagClear == "" && !cmd.Flags().Changed("swap") && !cmd.Flags().Changed("status") {
		return cmd.Help()
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments. Run 'git slot --help' for usage")
	}

	newBranch := flagCreate
	if flagBranch != "" {
		if newBranch != "" && newBranch != flagBranch {
			return fmt.Errorf("--create and --branch cannot specify different values")
		}
		newBranch = flagBranch
	}

	mgr, err := bootstrap()
	if err != nil {
		return err
	}

	if flagList {
		return runList(mgr, out)
	}
	if flagClear != "" {
		return runClear(mgr, flagClear, out)
	}
	if cmd.Flags().Changed("swap") {
		return runSwap(mgr, flagSwap, out)
	}
	if cmd.Flags().Changed("status") {
		return runStatus(mgr, flagStatus, out)
	}

	slotName := args[0]
	if len(args) == 2 {
		return runLoad(mgr, slotName, args[1], false, out)
	}
	if newBranch != "" {
		return runLoad(mgr, slotName, newBranch, true, out)
	}
	return runGetPath(mgr, slotName, out)
}

func bootstrap() (*slot.Manager, error) {
	detector := git.NewExecDetector("")
	repoRoot, err := detector.RepoRoot()
	if err != nil {
		return nil, err
	}

	var remote *git.RemoteInfo
	resolver := git.NewExecRemoteURLResolver(repoRoot)
	if rawURL, err := resolver.RemoteURL("origin"); err == nil {
		remote, _ = git.ParseRemoteURL(rawURL)
	}

	cfg, err := config.LoadConfig(config.LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		return nil, err
	}

	basePath, err := pathutil.ResolveSlotsBasePath(cfg, remote)
	if err != nil {
		return nil, err
	}

	wt := git.NewExecWorktree(repoRoot)
	return slot.NewManager(cfg, basePath, wt), nil
}

func runList(mgr *slot.Manager, out io.Writer) error {
	slots, err := mgr.List()
	if err != nil {
		return err
	}

	if len(slots) == 0 {
		_, _ = fmt.Fprintln(out, "No slots defined.")
		return nil
	}

	maxName := 0
	for _, s := range slots {
		if len(s.Name) > maxName {
			maxName = len(s.Name)
		}
	}

	for _, s := range slots {
		_, _ = fmt.Fprintln(out, formatSlotLine(s, maxName))
	}
	return nil
}

func formatSlotLine(s slot.Slot, nameWidth int) string {
	line := fmt.Sprintf("  %-*s  [%s]", nameWidth, s.Name, s.DisplayState())
	if s.State == slot.SlotActive {
		line += fmt.Sprintf("  %s  (%s)", s.Branch, s.HeadHash)
		if s.IsDirty {
			line += "  *dirty"
		}
	}
	return line
}

func runClear(mgr *slot.Manager, slotName string, out io.Writer) error {
	if err := mgr.Clear(slotName, slot.ClearOptions{Force: flagForce}); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "Slot '%s' is now empty.\n", slotName)
	return nil
}

func runLoad(mgr *slot.Manager, slotName, branchName string, createBranch bool, out io.Writer) error {
	err := mgr.Load(slotName, branchName, slot.LoadOptions{
		CreateBranch: createBranch,
		Force:        flagForce,
	})
	if err != nil {
		return err
	}
	path, _ := mgr.GetPath(slotName)
	_, _ = fmt.Fprintf(out, "Slot '%s' is ready.\n  Path: %s\n", slotName, path)
	return nil
}

func runGetPath(mgr *slot.Manager, slotName string, out io.Writer) error {
	path, err := mgr.GetPath(slotName)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, path)
	return nil
}

func runSwap(mgr *slot.Manager, swapArgs []string, out io.Writer) error {
	if len(swapArgs) != 2 {
		return fmt.Errorf("--swap requires exactly 2 slot names")
	}
	if err := mgr.Swap(swapArgs[0], swapArgs[1]); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "Swapped slots '%s' and '%s'.\n", swapArgs[0], swapArgs[1])
	return nil
}

func runStatus(mgr *slot.Manager, slotName string, out io.Writer) error {
	if slotName == "" {
		statuses, err := mgr.StatusAll()
		if err != nil {
			return err
		}
		for i, st := range statuses {
			if i > 0 {
				_, _ = fmt.Fprintln(out)
			}
			printSlotStatus(st, out)
		}
		return nil
	}
	st, err := mgr.Status(slotName)
	if err != nil {
		return err
	}
	printSlotStatus(*st, out)
	return nil
}

func printSlotStatus(st slot.SlotStatus, out io.Writer) {
	_, _ = fmt.Fprintf(out, "Slot:    %s\n", st.Name)
	_, _ = fmt.Fprintf(out, "State:   %s\n", st.DisplayState())
	if st.State == slot.SlotActive {
		_, _ = fmt.Fprintf(out, "Branch:  %s\n", st.Branch)
		_, _ = fmt.Fprintf(out, "HEAD:    %s", st.HeadHash)
		if st.CommitSubject != "" {
			_, _ = fmt.Fprintf(out, " (%s)", st.CommitSubject)
		}
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintf(out, "Path:    %s\n", st.Path)
		if len(st.Changes) > 0 {
			_, _ = fmt.Fprintln(out, "Changes:")
			for _, c := range st.Changes {
				_, _ = fmt.Fprintf(out, "  %s\n", c)
			}
		}
	}
}

func runInit(out io.Writer) error {
	detector := git.NewExecDetector("")
	repoRoot, _ := detector.RepoRoot()

	path, err := config.Init(config.InitOptions{
		Global: flagGlobal,
		Force:  flagForce,
	}, repoRoot)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "Created %s with template configuration.\n", path)
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(mapExitCode(err))
	}
}

func mapExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, config.ErrNoConfig) || errors.Is(err, config.ErrConfigParse) {
		return 2
	}
	if errors.Is(err, git.ErrNotInRepo) {
		return 3
	}
	return 1
}
