package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/spf13/cobra"
)

var swapCmd = &cobra.Command{
	Use:   "swap <slot-a> <slot-b>",
	Short: "Swap branches between two slots",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := bootstrap()
		if err != nil {
			return err
		}
		return runSwap(a.mgr, args, cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(swapCmd)
}

func runSwap(mgr slot.SlotManager, swapArgs []string, out io.Writer) error {
	if len(swapArgs) != 2 {
		return fmt.Errorf("--swap requires exactly 2 slot names")
	}
	if err := mgr.Swap(swapArgs[0], swapArgs[1]); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stderr, "Swapped slots '%s' and '%s'.\n", swapArgs[0], swapArgs[1])
	return nil
}
