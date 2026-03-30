package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <slot> [branch]",
	Short: "Set a branch into a slot, or print the slot path",
	Long: `Set a branch into a named slot, or print the slot's worktree path.

With one argument, prints the slot's worktree path (or launches a slot shell when
launch_shell is enabled and the slot already has a branch mounted — same as TUI).
With two arguments, mounts the specified branch into the slot.
Use -c/--create to create a new branch and mount it.
Use --pr to check out a GitHub PR branch into the slot.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runSet,
}

func init() {
	setCmd.Flags().StringP("create", "c", "", "Create a new branch and mount into slot")
	setCmd.Flags().StringP("branch", "b", "", "Alias for --create")
	setCmd.Flags().BoolP("force", "f", false, "Skip confirmation for destructive actions")
	setCmd.Flags().Bool("no-shell", false, "Suppress sub-shell launch even when launch_shell is enabled")
	setCmd.Flags().Int("pr", 0, "Check out the branch for a GitHub PR number into the slot")

	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	create, _ := cmd.Flags().GetString("create")
	branch, _ := cmd.Flags().GetString("branch")
	force, _ := cmd.Flags().GetBool("force")
	noShell, _ := cmd.Flags().GetBool("no-shell")
	prNumber, _ := cmd.Flags().GetInt("pr")

	if branch != "" {
		if create != "" && create != branch {
			return fmt.Errorf("--create and --branch cannot specify different values")
		}
		create = branch
	}

	if prNumber != 0 {
		if create != "" {
			return fmt.Errorf("--pr と --create は同時に使えません")
		}
		a, err := bootstrap()
		if err != nil {
			return err
		}
		return runPRCheckout(a, args[0], prNumber, mountOptions{force: force, noShell: noShell}, cmd.OutOrStdout())
	}

	a, err := bootstrap()
	if err != nil {
		return err
	}

	slotName := args[0]
	out := cmd.OutOrStdout()

	if len(args) == 2 {
		return runMount(a, slotName, args[1], mountOptions{force: force, noShell: noShell}, out)
	}
	if create != "" {
		return runMount(a, slotName, create, mountOptions{createBranch: true, force: force, noShell: noShell}, out)
	}
	return runGetPathOrLaunchShell(a, slotName, noShell, out)
}

func runPRCheckout(a *app, slotName string, prNumber int, opts mountOptions, out io.Writer) error {
	// 1. gh の存在確認
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI が必要です (https://cli.github.com)")
	}

	// 2. PR メタデータ取得
	ghCmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNumber),
		"--json", "headRefName,state")
	ghCmd.Dir = a.repoRoot
	ghOut, err := ghCmd.Output()
	if err != nil {
		return fmt.Errorf("PR #%d が見つかりません", prNumber)
	}

	var prInfo struct {
		HeadRefName string `json:"headRefName"`
		State       string `json:"state"`
	}
	if err := json.Unmarshal(ghOut, &prInfo); err != nil {
		return fmt.Errorf("gh 出力のパースに失敗: %w", err)
	}

	// 3. MERGED/CLOSED の場合は警告（処理は続行）
	if prInfo.State == "MERGED" || prInfo.State == "CLOSED" {
		fmt.Fprintf(os.Stderr, "Warning: PR #%d is %s\n", prNumber, prInfo.State)
	}

	// 4. fetch
	if err := a.wt.FetchBranch(prInfo.HeadRefName); err != nil {
		return fmt.Errorf("fetch に失敗しました: %w", err)
	}

	// 5. mount
	return runMount(a, slotName, prInfo.HeadRefName, opts, out)
}
