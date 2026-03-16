package cmd

import (
	"fmt"
	"io"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
)

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
