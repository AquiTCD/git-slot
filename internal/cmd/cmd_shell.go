package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/slotenv"
	"github.com/AquiTCD/git-slot/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

var ErrShellNested = errors.New("already inside a slot shell session; exit first, then re-run")

var shellCmd = &cobra.Command{
	Use:   "shell [slot]",
	Short: "Launch a sub-shell inside a slot with GSL_* environment variables",
	Long: `Launch a new shell session inside the specified slot's worktree directory.

The shell inherits the current environment plus slot-specific variables
(GSL_SLOT_NAME, GSL_SLOT_PATH, GSL_BRANCH, GSL_REPO_ROOT, GSL_SHELL_SESSION)
and any user-defined env from the slot configuration.

If no slot is specified and stdin is a TTY, opens interactive TUI for selection.

Nesting is not allowed: running this inside an existing slot shell will fail.
Use 'exit' to leave the current slot shell first.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShell,
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

func runShell(cmd *cobra.Command, args []string) error {
	if err := checkShellNesting(); err != nil {
		return err
	}

	a, err := bootstrap()
	if err != nil {
		return err
	}

	var slotName string
	if len(args) == 1 {
		slotName = args[0]
	} else {
		name, err := selectSlotInteractive(a)
		if err != nil {
			return err
		}
		if name == "" {
			return nil
		}
		slotName = name
	}

	return launchSlotShell(a, slotName)
}

func launchSlotShell(a *app, slotName string) error {
	st, err := a.mgr.Status(slotName)
	if err != nil {
		return err
	}
	if st.State == slot.SlotEmpty {
		return fmt.Errorf("slot '%s' is empty; mount a branch first with 'git slot set'", slotName)
	}

	printHintSlotShell()

	slotDef := a.cfg.FindSlot(slotName)
	var userEnv map[string]string
	if slotDef != nil {
		userEnv = slotDef.Env
	}

	info := slotenv.SlotInfo{
		SlotName: slotName,
		SlotPath: st.Path,
		Branch:   st.Branch,
		RepoRoot: a.repoRoot,
	}
	slotVars := slotenv.BuildSlotEnv(info, userEnv)
	mergedEnv := slotenv.MergeEnvWithOS(os.Environ(), slotVars)

	return execShell(st.Path, mergedEnv)
}

var execShellFunc = execShellDefault

func execShell(dir string, env []string) error {
	return execShellFunc(dir, env)
}

func execShellDefault(dir string, env []string) error {
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("slot shell: chdir to %s: %w", dir, err)
	}
	env = envWithPWD(env, dir)

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	// When git-slot runs under command substitution (e.g. the gsl wrapper's
	// result=$(command git-slot set ...)), stdout is a pipe. syscall.Exec would
	// leave the interactive shell with a non-TTY stdout, so prompts never show
	// and the shell often exits immediately. Rebind to the controlling terminal.
	_ = attachStdioToControllingTTY()

	return syscall.Exec(shell, []string{shell}, env)
}

// envWithPWD drops any inherited PWD and sets PWD to dir so the exec'd shell
// matches the process working directory after Chdir.
func envWithPWD(env []string, dir string) []string {
	pwd := "PWD=" + dir
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if strings.HasPrefix(e, "PWD=") {
			continue
		}
		out = append(out, e)
	}
	out = append(out, pwd)
	return out
}

// attachStdioToControllingTTY redirects stdout/stderr to the controlling tty when
// they are not terminals. Best-effort: failure is ignored so direct invocations
// and headless environments keep previous behavior.
func attachStdioToControllingTTY() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	needOut := !term.IsTerminal(int(os.Stdout.Fd()))
	needErr := !term.IsTerminal(int(os.Stderr.Fd()))
	if !needOut && !needErr {
		return nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer func() { _ = tty.Close() }()

	tfd := int(tty.Fd())
	if needOut {
		if err := unix.Dup2(tfd, int(os.Stdout.Fd())); err != nil {
			return err
		}
	}
	if needErr {
		if err := unix.Dup2(tfd, int(os.Stderr.Fd())); err != nil {
			return err
		}
	}
	return nil
}

func selectSlotInteractive(a *app) (string, error) {
	slots, err := a.mgr.List()
	if err != nil {
		return "", err
	}

	activeSlots := make([]slot.Slot, 0, len(slots))
	for _, s := range slots {
		if s.State != slot.SlotEmpty {
			activeSlots = append(activeSlots, s)
		}
	}
	if len(activeSlots) == 0 {
		return "", fmt.Errorf("no active slots; mount a branch first with 'git slot set'")
	}

	if !tui.IsTTY(os.Stdin) {
		return "", fmt.Errorf("interactive mode requires a TTY")
	}

	var branches []string
	noColor := tui.IsNoColor()
	model := tui.NewInteractiveModel(activeSlots, branches, noColor)

	m, aborted, err := runTUI[tui.Model](model)
	if err != nil {
		return "", err
	}
	if aborted {
		return "", nil
	}

	result, ok := m.GetResult()
	if !ok {
		return "", nil
	}

	return result.SlotName, nil
}
