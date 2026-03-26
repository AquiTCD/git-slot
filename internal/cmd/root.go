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
Set branches into slots, clear them, and more.

Without arguments, opens interactive TUI for slot selection.`,
	SilenceUsage:          true,
	SilenceErrors:         true,
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	RunE:                  run,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "Print version information")
	rootCmd.SetUsageTemplate(`Usage:
  {{.CommandPath}} [command]{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
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
		return runInteractive(a, false, false, cmd.OutOrStdout())
	}
	return cmd.Help()
}

type app struct {
	mgr      slot.SlotManager
	cfg      *config.Config
	repoRoot string
	wt       git.Worktree
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

	var repoName string
	if remote != nil {
		repoName = remote.Repo
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
		mgr:      slot.NewManager(cfg, basePath, repoName, wt),
		cfg:      cfg,
		repoRoot: repoRoot,
		wt:       wt,
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
