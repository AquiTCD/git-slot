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
		if noColor {
			parts = append(parts, icon)
		} else {
			parts = append(parts, StyleIcon.Render(icon))
		}
	}

	name := fmt.Sprintf("%-*s", nameWidth, s.Name)
	if noColor {
		parts = append(parts, name)
	} else {
		parts = append(parts, StyleName.Render(name))
	}

	state := s.DisplayState()
	stateTag := fmt.Sprintf("[%s]", state)
	if noColor {
		parts = append(parts, stateTag)
	} else {
		parts = append(parts, StateStyle(state).Render(stateTag))
	}

	if s.State == slot.SlotActive {
		if noColor {
			parts = append(parts, s.Branch)
			parts = append(parts, fmt.Sprintf("(%s)", s.HeadHash))
		} else {
			parts = append(parts, StyleBranch.Render(s.Branch))
			parts = append(parts, StyleHash.Render(fmt.Sprintf("(%s)", s.HeadHash)))
		}
		if s.IsDirty {
			if noColor {
				parts = append(parts, "*dirty")
			} else {
				parts = append(parts, StyleDirtyMark.Render("*dirty"))
			}
		}
	}

	return "  " + lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(parts, "  "))
}
