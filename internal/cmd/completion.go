package cmd

import "github.com/spf13/cobra"

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

func init() {
	rootCmd.AddCommand(completionCmd)
}
