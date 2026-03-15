package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/errutil"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
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
	flagEject   bool
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
  git slot -e, --eject           Print the repository root path (use with gsl to cd back)
  git slot --status [slot]       Show detailed slot status
  git slot --init                Generate a template config file`,
	SilenceUsage:          true,
	SilenceErrors:         true,
	Args:                  cobra.ArbitraryArgs,
	TraverseChildren:      true,
	DisableFlagsInUseLine: true,
	RunE:                  run,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagList, "list", "l", false, "List all slots and their status")
	rootCmd.Flags().StringVarP(&flagClear, "clear", "d", "", "Clear (remove) a slot's worktree")
	rootCmd.Flags().StringSliceVarP(&flagSwap, "swap", "s", nil, "Swap branches between two slots")
	rootCmd.Flags().StringVar(&flagStatus, "status", "", "Show detailed slot status")
	rootCmd.Flags().BoolVar(&flagInit, "init", false, "Generate a template config file")
	rootCmd.Flags().BoolVarP(&flagGlobal, "global", "g", false, "Used with --init to generate global config")
	rootCmd.Flags().StringVarP(&flagCreate, "create", "c", "", "Create a new branch and load into slot")
	rootCmd.Flags().StringVarP(&flagBranch, "branch", "b", "", "Alias for --create")
	rootCmd.Flags().BoolVar(&flagForce, "force", false, "Skip confirmation for destructive actions")
	rootCmd.Flags().BoolVarP(&flagEject, "eject", "e", false, "Print the repository root path (use with gsl to cd back)")
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

	if flagEject {
		a, err := bootstrap()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(out, a.repoRoot)
		return nil
	}

	if len(args) == 0 && !flagList && flagClear == "" && !cmd.Flags().Changed("swap") && !cmd.Flags().Changed("status") {
		if tui.IsTTY(os.Stdin) {
			a, err := bootstrap()
			if err != nil {
				return err
			}
			return runInteractive(a, cmd.OutOrStdout())
		}
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

	a, err := bootstrap()
	if err != nil {
		return err
	}

	if flagList {
		return runList(a.mgr, out, flagJSON)
	}
	if flagClear != "" {
		return runClear(a, flagClear, out)
	}
	if cmd.Flags().Changed("swap") {
		return runSwap(a.mgr, flagSwap, out)
	}
	if cmd.Flags().Changed("status") {
		return runStatus(a.mgr, flagStatus, out, flagJSON)
	}

	slotName := args[0]
	if len(args) == 2 {
		return runLoad(a, slotName, args[1], false, out)
	}
	if newBranch != "" {
		return runLoad(a, slotName, newBranch, true, out)
	}
	return runGetPath(a.mgr, slotName, out)
}

type app struct {
	mgr      *slot.Manager
	cfg      *config.Config
	basePath string
	repoRoot string
}

func bootstrap() (*app, error) {
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
	return &app{
		mgr:      slot.NewManager(cfg, basePath, wt),
		cfg:      cfg,
		basePath: basePath,
		repoRoot: repoRoot,
	}, nil
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
	var ex errutil.ExitError
	if errors.As(err, &ex) {
		return ex.ExitCode()
	}
	return 1
}
