package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// IconCellWidth is how many terminal cells we reserve for slot icons so the name
// and following fields start at a stable column regardless of emoji width quirks.
const IconCellWidth = 3

// displayCellWidth is the estimated terminal cell count for a string that may
// contain ANSI (e.g. lipgloss); escapes are stripped before measuring.
func displayCellWidth(s string) int {
	if s == "" {
		return 0
	}
	return terminalDisplayWidthPlain(ansi.Strip(s))
}

// terminalDisplayWidthPlain estimates terminal cell width for plain text (no ANSI).
// It is grapheme-aware and bumps widths that wcwidth-style tables often undercount
// (keycap sequences, regional-indicator flags).
func terminalDisplayWidthPlain(s string) int {
	if s == "" {
		return 0
	}
	gr := uniseg.NewGraphemes(s)
	n := 0
	for gr.Next() {
		n += clusterTermCells(gr.Str())
	}
	return n
}

func clusterTermCells(cl string) int {
	if cl == "" {
		return 0
	}
	// Single-cell box-drawing used for split-pane separator; runewidth often returns 2
	// (ambiguous width) while ansi.StringWidth and common terminals use 1.
	if rs := []rune(cl); len(rs) == 1 {
		switch rs[0] {
		case '\u2500', '\u2502', '\u250c', '\u2510', '\u2514', '\u2518', '\u251c', '\u2524', '\u252c', '\u2534', '\u253c':
			return 1
		}
	}
	// Misc symbols often emoji-presented as 2 cells; runewidth/ansi may report 1 (e.g. ⚡).
	if rs := []rune(cl); len(rs) == 1 {
		switch rs[0] {
		case '\u26a1', '\u26a0', '\u26d4', '\u26bd', '\u26be', '\u26f3', '\u26f5', '\u26fa', '\u26fd':
			return max(2, runewidth.StringWidth(cl))
		}
	}
	// Digit / * / # + VS-16 + U+20E3 (enclosing keycap): runewidth often returns 1.
	if strings.ContainsRune(cl, '\u20e3') {
		return max(2, runewidth.StringWidth(cl))
	}
	// Regional-indicator pairs (e.g. 🇯🇵): runewidth often returns 1 for the cluster.
	if isRegionalIndicatorSequence(cl) {
		return max(2, runewidth.StringWidth(cl))
	}
	w := runewidth.StringWidth(cl)
	if w < 1 {
		return 1
	}
	return w
}

func isRegionalIndicatorSequence(cl string) bool {
	rs := []rune(cl)
	if len(rs) < 2 || len(rs)%2 != 0 {
		return false
	}
	for _, r := range rs {
		if r < 0x1F1E6 || r > 0x1F1FF {
			return false
		}
	}
	return true
}

func truncatePlainIconToMax(icon string, maxCells int) string {
	if maxCells <= 0 {
		return ""
	}
	if terminalDisplayWidthPlain(icon) <= maxCells {
		return icon
	}
	var b strings.Builder
	w := 0
	gr := uniseg.NewGraphemes(icon)
	for gr.Next() {
		cl := gr.Str()
		cw := clusterTermCells(cl)
		if w+cw > maxCells {
			break
		}
		b.WriteString(cl)
		w += cw
	}
	return b.String()
}

// formatIconGutter lays out the icon in a fixed IconCellWidth-wide field (padded
// with spaces or truncated). Empty icon still consumes IconCellWidth cells.
// Width budgeting uses terminalDisplayWidthPlain so keycaps and flags align
// better with common emulators than ansi.StringWidth alone.
func formatIconGutter(icon string, noColor bool, rowHighlight bool) string {
	useRow := rowHighlight && !noColor
	if icon == "" {
		if useRow {
			return StyleFzfRowBG.Render(strings.Repeat(" ", IconCellWidth))
		}
		return strings.Repeat(" ", IconCellWidth)
	}
	fit := truncatePlainIconToMax(icon, IconCellWidth)
	var rendered string
	if useRow {
		rendered = renderOnFzfRowBG(StyleIcon, fit)
		pad := IconCellWidth - terminalDisplayWidthPlain(fit)
		if pad > 0 {
			rendered += StyleFzfRowBG.Render(strings.Repeat(" ", pad))
		}
	} else {
		rendered = render(StyleIcon, fit, noColor)
		pad := IconCellWidth - terminalDisplayWidthPlain(fit)
		if pad > 0 {
			rendered += strings.Repeat(" ", pad)
		}
	}
	// Styles should be zero-width; keep ansi-based check as a safety net.
	if ansi.StringWidth(rendered) > IconCellWidth {
		rendered = ansi.Truncate(rendered, IconCellWidth, "")
	}
	return rendered
}
