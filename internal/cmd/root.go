package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/errutil"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
	"github.com/spf13/cobra"
)

var flagVersion bool

var rootCmd = &cobra.Command{
	Use:   "git-slot [command]",
	Short: "Manage git worktrees as fixed slots",
	Long: `git-slot manages git worktrees as fixed, named slots defined in TOML configuration.
Set branches into slots, clear them, swap between them, and more.

Subcommands:
  git slot set <slot> [branch]   Set a branch into a slot, or print slot path
  git slot list                  List all slots and their status
  git slot clear <slot>          Clear (remove) a slot's worktree
  git slot swap <A> <B>          Swap branches between two slots
  git slot status [slot]         Show detailed slot status
  git slot init                  Generate a template config file
  git slot hook                  Open interactive TUI to setup post-mount hooks
  git slot root                  Print the repository root path

Without arguments, opens interactive TUI for slot selection.`,
	SilenceUsage:          true,
	SilenceErrors:         true,
	Args:                  cobra.NoArgs,
	TraverseChildren:      true,
	DisableFlagsInUseLine: true,
	RunE:                  run,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "Print version information")
}

func run(cmd *cobra.Command, _ []string) error {
	if flagVersion {
		printVersion(cmd.OutOrStdout())
		return nil
	}

	if tui.IsTTY(os.Stdin) {
		a, err := bootstrap()
		if err != nil {
			return err
		}
		return runInteractive(a, false, cmd.OutOrStdout())
	}
	return cmd.Help()
}

type app struct {
	mgr      slot.SlotManager
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
