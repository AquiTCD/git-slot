package cmd

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/slotenv"
	"github.com/AquiTCD/git-slot/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
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

func checkShellNesting() error {
	if os.Getenv("GSL_SHELL_SESSION") != "" {
		return ErrShellNested
	}
	return nil
}

func launchSlotShell(a *app, slotName string) error {
	slotPath := pathutil.ResolveSlotPath(a.basePath, slotName)

	if _, err := a.mgr.GetPath(slotName); err != nil {
		return fmt.Errorf("slot '%s' is empty; mount a branch first with 'git slot set'", slotName)
	}

	slots, err := a.mgr.List()
	if err != nil {
		return err
	}
	var branch string
	for _, s := range slots {
		if s.Name == slotName {
			branch = s.Branch
			break
		}
	}

	slotDef := findSlotDef(a.cfg, slotName)
	var userEnv map[string]string
	if slotDef != nil {
		userEnv = slotDef.Env
	}

	info := slotenv.SlotInfo{
		SlotName: slotName,
		SlotPath: slotPath,
		Branch:   branch,
		RepoRoot: a.repoRoot,
	}
	slotVars := slotenv.BuildSlotEnv(info, userEnv)
	mergedEnv := slotenv.MergeEnvWithOS(os.Environ(), slotVars)

	return execShell(slotPath, mergedEnv)
}

func findSlotDef(cfg *config.Config, name string) *config.SlotDefinition {
	for i := range cfg.Slots {
		if cfg.Slots[i].Name == name {
			return &cfg.Slots[i]
		}
	}
	return nil
}

var execShellFunc = execShellDefault

func execShell(dir string, env []string) error {
	return execShellFunc(dir, env)
}

func execShellDefault(dir string, env []string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	return syscall.Exec(shell, []string{shell}, env)
}

func selectSlotInteractive(a *app) (string, error) {
	slots, err := a.mgr.List()
	if err != nil {
		return "", err
	}

	activeSlots := make([]slot.Slot, 0)
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

	branches := []string{}
	noColor := tui.IsNoColor()
	model := tui.NewInteractiveModel(activeSlots, branches, noColor)

	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("interactive mode: %w", err)
	}

	m := finalModel.(tui.Model)
	if m.Aborted() {
		return "", nil
	}

	result, ok := m.GetResult()
	if !ok {
		return "", nil
	}

	return result.SlotName, nil
}
