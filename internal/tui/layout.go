package tui

import (
	"math"
	"strings"
)

// goldenRatio is φ (1.618…). Left column (slots + branches) is the larger segment:
// left : right ≈ φ : 1 over (inner − separator).
const goldenRatio = 1.6180339887498948482045868364

// slotSelectLayout holds column sizes for the interactive slot-select grid.
// Invariant: Inner == Left + 1 + Right (the single-cell │ separator).
type slotSelectLayout struct {
	Inner, Left, Right, LeftContent, RightContent int
}

// computeSlotSelectLayout derives column widths from the terminal width.
// termWidth <= 0 uses 88 (Bubble Tea default before the first WindowSizeMsg).
func computeSlotSelectLayout(termWidth int) slotSelectLayout {
	if termWidth <= 0 {
		termWidth = 88
	}
	const leftPad = 1
	inner := termWidth - 4  // RoundedBorder l+r + horizontal Padding l+r
	splittable := inner - 1 // room for │ between columns
	if splittable < 2 {
		splittable = 2
	}
	right := int(math.Round(float64(splittable) / (1 + goldenRatio)))
	left := splittable - right
	if right < 10 {
		right = 10
		left = splittable - right
	}
	if left < 1 {
		left = 1
		right = splittable - left
	}
	leftContent := left - leftPad
	if leftContent < 1 {
		leftContent = 1
	}
	rightContent := right - leftPad
	if rightContent < 8 {
		rightContent = 8
	}
	return slotSelectLayout{
		Inner:        inner,
		Left:         left,
		Right:        right,
		LeftContent:  leftContent,
		RightContent: rightContent,
	}
}

// filterFieldWidth returns bubbles textinput.Width for the filter row for this terminal.
func filterFieldWidth(termWidth int) int {
	lay := computeSlotSelectLayout(termWidth)
	return max(8, lay.LeftContent)
}

// joinSplitPaneRows pairs left and right column blocks line-by-line, padding the shorter
// side with spaces so each row is exactly Left + ansiWidth(sep) + Right cells wide.
func joinSplitPaneRows(leftBlock, rightBlock, sep string, lay slotSelectLayout) []string {
	leftLines := strings.Split(leftBlock, "\n")
	rightLines := strings.Split(rightBlock, "\n")
	height := max(len(leftLines), len(rightLines))
	emptyLeft := strings.Repeat(" ", lay.Left)
	emptyRight := strings.Repeat(" ", lay.Right)
	for len(leftLines) < height {
		leftLines = append(leftLines, emptyLeft)
	}
	for len(rightLines) < height {
		rightLines = append(rightLines, emptyRight)
	}
	rows := make([]string, height)
	for i := range rows {
		rows[i] = leftLines[i] + sep + rightLines[i]
	}
	return rows
}
