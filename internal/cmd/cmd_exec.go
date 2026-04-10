package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/AquiTCD/git-slot/internal/errutil"
	"github.com/AquiTCD/git-slot/internal/slotenv"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Run a command in a slot worktree with slot environment variables",
	Long: `Runs a command with the slot's worktree as working directory and with GSL_* variables
plus any [[slots.env]] entries from configuration. GSL_SHELL_SESSION is not set (unlike
git slot shell).

Use '--' before the command. If no slot name is given, the slot is inferred from the
current directory (same rules as git slot which).

From an interactive slot shell (GSL_SHELL_SESSION=1), only the current slot may be used.

Examples:
  git slot exec -- npm test
  git slot exec main-work -- make build`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	RunE:               runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

var execRunHook = func(c *exec.Cmd) error {
	return c.Run()
}

func runExec(cmd *cobra.Command, _ []string) error {
	if execWantsHelp(os.Args) {
		return cmd.Help()
	}
	a, err := bootstrap()
	if err != nil {
		return err
	}
	return runExecFromApp(a, os.Args)
}

// execWantsHelp is true for `git slot exec --help` / `-h` before `--` (not npm --help).
func execWantsHelp(argv []string) bool {
	seg, ok := execArgumentSegment(argv)
	if !ok {
		return false
	}
	for _, a := range seg {
		if a == "--" {
			return false
		}
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}

func runExecFromApp(a *app, argv []string) error {
	explicit, slotArg, cmdArgs, err := parseExecArgv(argv, a.cfg)
	if err != nil {
		return err
	}
	var slotName string
	if explicit {
		slotName = slotArg
	} else {
		slotName, err = a.slotNameFromGitCWD()
		if err != nil {
			return exitIfNotSlotWorktree(err, "not inside a configured git-slot worktree; specify a slot before '--'")
		}
	}
	if err := checkExecAllowedInSlotShell(explicit, slotArg, slotName); err != nil {
		return err
	}
	st, info, userEnv, err := loadActiveSlotContext(a, slotName)
	if err != nil {
		return err
	}
	merged := slotenv.MergeEnvWithOS(os.Environ(), slotenv.BuildSlotExecEnv(info, userEnv))
	exeCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	exeCmd.Dir = st.Path
	exeCmd.Env = merged
	exeCmd.Stdin = os.Stdin
	exeCmd.Stdout = os.Stdout
	exeCmd.Stderr = os.Stderr
	return runExecExit(execRunHook(exeCmd))
}

func runExecExit(err error) error {
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return errutil.NewExitError(exitErr.Error(), exitErr.ExitCode())
	}
	return err
}
