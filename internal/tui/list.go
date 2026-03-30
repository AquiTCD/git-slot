package tui

import (
	"fmt"
	"strings"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/charmbracelet/lipgloss"
)

func RenderSlotList(slots []slot.Slot, noColor bool) string {
	if len(slots) == 0 {
		return "No slots defined."
	}

	maxName := 0
	hasIcons := false
	for _, s := range slots {
		if len(s.Name) > maxName {
			maxName = len(s.Name)
		}
		if s.Icon != "" {
			hasIcons = true
		}
	}

	var b strings.Builder
	for i, s := range slots {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(renderSlotLine(s, maxName, hasIcons, noColor))
	}
	return b.String()
}

func renderSlotLine(s slot.Slot, nameWidth int, hasIcons, noColor bool) string {
	var parts []string

	if hasIcons {
		icon := "  "
		if s.Icon != "" {
			icon = s.Icon
		}
		parts = append(parts, render(StyleIcon, icon, noColor))
	}

	name := fmt.Sprintf("%-*s", nameWidth, s.Name)
	parts = append(parts, render(StyleName, name, noColor))

	state := s.DisplayState()
	stateTag := fmt.Sprintf("[%s]", state)
	parts = append(parts, render(StateStyle(state), stateTag, noColor))

	if s.State == slot.SlotActive {
		parts = append(parts, render(StyleBranch, s.Branch, noColor))
		parts = append(parts, render(StyleHash, fmt.Sprintf("(%s)", s.HeadHash), noColor))

		var dirtyMark string
		if s.DirtyCount > 0 {
			dirtyMark = fmt.Sprintf("*%d", s.DirtyCount)
		} else {
			dirtyMark = "clean"
		}
		parts = append(parts, render(StyleDirtyMark, dirtyMark, noColor))

		if s.HasUpstream {
			syncMark := fmt.Sprintf("↑%d ↓%d", s.AheadCount, s.BehindCount)
			parts = append(parts, render(StyleAheadMark, syncMark, noColor))
		}
	}

	return "  " + lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(parts, "  "))
}
