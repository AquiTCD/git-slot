package tui

import (
	"fmt"
	"strings"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func maxSlotNameLen(slots []slot.Slot) int {
	maxName := 0
	for _, s := range slots {
		if len(s.Name) > maxName {
			maxName = len(s.Name)
		}
	}
	return maxName
}

func renderSlotHeadField(useRow bool, style lipgloss.Style, text string, noColor bool) string {
	if useRow {
		return renderOnFzfRowBG(style, text)
	}
	return render(style, text, noColor)
}

func RenderSlotList(slots []slot.Slot, noColor bool) string {
	if len(slots) == 0 {
		return "No slots defined."
	}

	maxName := maxSlotNameLen(slots)

	var b strings.Builder
	for i, s := range slots {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(renderSlotLine(s, maxName, noColor))
	}
	return b.String()
}

// slotLineHeadTailStrings returns the line in two segments: through branch name (for
// active slots), then hash / dirty / sync markers. tail is empty for non-active slots.
// When fzfHeadRowBG is true and color is on, each head field is rendered with the fzf
// row background so inner lipgloss resets do not clear the highlight after the icon.
func slotLineHeadTailStrings(s slot.Slot, nameWidth int, noColor bool, fzfHeadRowBG bool) (head, tail string) {
	useRow := fzfHeadRowBG && !noColor
	sep := "  "
	if useRow {
		sep = StyleFzfRowBG.Render("  ")
	}
	var headParts []string
	headParts = append(headParts, formatIconGutter(s.Icon, noColor, fzfHeadRowBG))
	name := fmt.Sprintf("%-*s", nameWidth, s.Name)
	headParts = append(headParts, renderSlotHeadField(useRow, StyleName, name, noColor))
	state := s.DisplayState()
	stateTag := fmt.Sprintf("[%s]", state)
	headParts = append(headParts, renderSlotHeadField(useRow, StateStyle(state), stateTag, noColor))
	if s.State != slot.SlotActive {
		return strings.Join(headParts, sep), ""
	}
	headParts = append(headParts, renderSlotHeadField(useRow, StyleBranch, s.Branch, noColor))
	head = strings.Join(headParts, sep)

	var tailParts []string
	tailParts = append(tailParts, render(StyleHash, fmt.Sprintf("(%s)", s.HeadHash), noColor))
	var dirtyMark string
	if s.DirtyCount > 0 {
		dirtyMark = fmt.Sprintf("*%d", s.DirtyCount)
	} else {
		dirtyMark = "clean"
	}
	tailParts = append(tailParts, render(StyleDirtyMark, dirtyMark, noColor))
	if s.HasUpstream {
		syncMark := fmt.Sprintf("↑%d ↓%d", s.AheadCount, s.BehindCount)
		tailParts = append(tailParts, render(StyleAheadMark, syncMark, noColor))
	}
	tail = strings.Join(tailParts, "  ")
	return head, tail
}

func slotLineBody(s slot.Slot, nameWidth int, noColor bool) string {
	head, tail := slotLineHeadTailStrings(s, nameWidth, noColor, false)
	out := head
	if tail != "" {
		out = head + "  " + tail
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, out)
}

func renderSlotLine(s slot.Slot, nameWidth int, noColor bool) string {
	return "  " + slotLineBody(s, nameWidth, noColor)
}

// RenderSlotFilterDivider draws an fzf-style rule: horizontal dashes plus " shown/total ".
func RenderSlotFilterDivider(termWidth, shown, total int, noColor bool) string {
	suffix := fmt.Sprintf(" %d/%d", shown, total)
	tw := termWidth
	if tw <= 0 {
		tw = 80
	}
	dashCount := tw - ansi.StringWidth(suffix)
	if dashCount < 4 {
		dashCount = 4
	}
	line := strings.Repeat("─", dashCount) + suffix
	if noColor {
		return line
	}
	return render(StyleFzfDivider, line, false)
}

// RenderInteractiveSlotLine renders one slot row like git slot list, with fzf-like
// selection when isSelected and color is on: magenta bar + background only through the
// branch segment (hash / dirty / sync stay un-highlighted). No terminal-width padding.
func RenderInteractiveSlotLine(s slot.Slot, nameWidth int, isSelected, noColor bool) string {
	body := slotLineBody(s, nameWidth, noColor)
	if isSelected && !noColor {
		bar := StyleFzfBar.Render(" ")
		head, tail := slotLineHeadTailStrings(s, nameWidth, noColor, true)
		// Unselected rows use two spaces before body; bar (1) + space (1) keeps alignment.
		prefix := StyleFzfRowBG.Render(" ")
		if tail == "" {
			return lipgloss.JoinHorizontal(lipgloss.Top, bar, prefix, head)
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, bar, prefix, head, "  "+tail)
	}
	if isSelected && noColor {
		return "> " + body
	}
	return "  " + body
}
