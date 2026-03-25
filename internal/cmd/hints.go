package cmd

import (
	"fmt"
	"os"
)

func printHintSlotShell() {
	_, _ = fmt.Fprintln(os.Stderr, "git-slot: Starting a slot shell. Type 'exit' to leave.")
}
