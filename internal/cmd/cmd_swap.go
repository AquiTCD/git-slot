package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/slot"
)

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
