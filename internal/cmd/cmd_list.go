package cmd

import (
	"fmt"
	"io"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all slots and their status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		useJSON, _ := cmd.Flags().GetBool("json")

		a, err := bootstrap()
		if err != nil {
			return err
		}
		return runList(a.mgr, cmd.OutOrStdout(), useJSON)
	},
}

func init() {
	listCmd.Flags().Bool("json", false, "Output in JSON format")

	rootCmd.AddCommand(listCmd)
}

func runList(mgr slot.SlotManager, out io.Writer, useJSON bool) error {
	slots, err := mgr.List()
	if err != nil {
		return err
	}

	if useJSON {
		items := make([]jsonSlot, len(slots))
		for i, s := range slots {
			items[i] = slotToJSON(s)
		}
		return writeJSON(out, jsonSlotList{Slots: items})
	}

	noColor := tui.IsNoColor()
	_, _ = fmt.Fprintln(out, tui.RenderSlotList(slots, noColor))
	return nil
}
