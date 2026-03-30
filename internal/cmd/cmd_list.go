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
	if useJSON {
		statuses, err := mgr.StatusAll()
		if err != nil {
			return err
		}
		items := make([]jsonSlot, len(statuses))
		for i, s := range statuses {
			items[i] = statusToJSONSlot(s)
		}
		return writeJSON(out, jsonSlotList{Slots: items})
	}

	slots, err := mgr.List()
	if err != nil {
		return err
	}
	noColor := tui.IsNoColor()
	_, _ = fmt.Fprintln(out, tui.RenderSlotList(slots, noColor))
	return nil
}
