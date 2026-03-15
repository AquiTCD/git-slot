package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for git-slot.

To load completions:

Bash:
  $ source <(git-slot completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ git-slot completion bash > /etc/bash_completion.d/git-slot
  # macOS:
  $ git-slot completion bash > $(brew --prefix)/etc/bash_completion.d/git-slot

Zsh:
  $ source <(git-slot completion zsh)
  # To load completions for each session, execute once:
  $ git-slot completion zsh > "${fpath[1]}/_git-slot"

Fish:
  $ git-slot completion fish | source
  # To load completions for each session, execute once:
  $ git-slot completion fish > ~/.config/fish/completions/git-slot.fish

PowerShell:
  PS> git-slot completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(out)
		case "zsh":
			return cmd.Root().GenZshCompletion(out)
		case "fish":
			return cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}

var wrapperCmd = &cobra.Command{
	Use:   "wrapper [bash|zsh|fish]",
	Short: "Generate gsl shell wrapper function",
	Long: `Generate a shell wrapper function "gsl" that calls git-slot and
automatically cd's into the slot directory on success.

To enable this, add the following to your shell config file:

Bash / Zsh (~/.zshrc or ~/.bashrc):
  eval "$(git-slot wrapper zsh)"

Fish (~/.config/fish/config.fish):
  git-slot wrapper fish | source
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash", "zsh":
			return writeShWrapper(out)
		case "fish":
			return writeFishWrapper(out)
		}
		return nil
	},
}

func writeShWrapper(w io.Writer) error {
	_, err := fmt.Fprintf(w, `gsl() {
  local result
  # Force color output even when captured in a variable,
  # but only if NO_COLOR is not set.
  if [ -z "$NO_COLOR" ]; then
    result=$(CLICOLOR_FORCE=1 command git-slot "$@" </dev/tty)
  else
    result=$(command git-slot "$@" </dev/tty)
  fi
  local rc=$?
  if [ $rc -eq 0 ]; then
    if [ -n "$result" ] && [ -d "$result" ]; then
      cd "$result" || return 1
    elif [ -n "$result" ]; then
      printf "%%s\n" "$result"
    fi
  fi
  return $rc
}
`)
	return err
}

func writeFishWrapper(w io.Writer) error {
	_, err := fmt.Fprintf(w, `function gsl
  set -l result
  if not set -q NO_COLOR
    set result (env CLICOLOR_FORCE=1 command git-slot $argv </dev/tty)
  else
    set result (command git-slot $argv </dev/tty)
  end
  set -l rc $status
  if test $rc -eq 0
    if test -n "$result"; and test -d "$result[1]"
      cd "$result[1]"
    else if test -n "$result"
      printf "%%s\n" $result
    end
  end
  return $rc
end
`)
	return err
}

func init() {
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(wrapperCmd)
}
