package tui

// filterFieldWidth returns bubbles textinput.Width for the filter row for this terminal.
// termWidth <= 0 uses 88 (Bubble Tea default before the first WindowSizeMsg).
func filterFieldWidth(termWidth int) int {
	if termWidth <= 0 {
		termWidth = 88
	}
	// Leave modest margin so the filter field does not touch the terminal edge.
	inner := max(8, termWidth-6)
	return inner
}
