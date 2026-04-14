package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestTerminalDisplayWidthPlain_Keycaps(t *testing.T) {
	assert.Equal(t, 2, terminalDisplayWidthPlain("2️⃣"))
	assert.Equal(t, 1, terminalDisplayWidthPlain("2"))
}

func TestTerminalDisplayWidthPlain_Flag(t *testing.T) {
	assert.Equal(t, 2, terminalDisplayWidthPlain("🇯🇵"))
}

func TestTerminalDisplayWidthPlain_Seedling(t *testing.T) {
	assert.Equal(t, 2, terminalDisplayWidthPlain("🌱"))
}

func TestFormatIconGutter_Empty(t *testing.T) {
	got := formatIconGutter("", true, false)
	assert.Equal(t, IconCellWidth, ansi.StringWidth(got))
	assert.Equal(t, IconCellWidth, terminalDisplayWidthPlain(got))
	assert.Equal(t, "   ", got)
}

func TestFormatIconGutter_ShortASCII(t *testing.T) {
	got := formatIconGutter("x", true, false)
	assert.Equal(t, IconCellWidth, ansi.StringWidth(got))
	assert.Equal(t, IconCellWidth, terminalDisplayWidthPlain(got))
	assert.Equal(t, "x  ", got)
}

func TestFormatIconGutter_NoColorEmoji(t *testing.T) {
	got := formatIconGutter("🌱", true, false)
	assert.Equal(t, IconCellWidth, terminalDisplayWidthPlain(got))
}

func TestFormatIconGutter_KeycapFitsThreeCells(t *testing.T) {
	got := formatIconGutter("2️⃣", true, false)
	assert.Equal(t, IconCellWidth, terminalDisplayWidthPlain(got))
	assert.Contains(t, got, "2")
}

func TestFormatIconGutter_FlagFitsThreeCells(t *testing.T) {
	got := formatIconGutter("🇯🇵", true, false)
	assert.Equal(t, IconCellWidth, terminalDisplayWidthPlain(got))
}

func TestFormatIconGutter_RowHighlightFitsIconColumn(t *testing.T) {
	got := formatIconGutter("🌱", false, true)
	assert.Equal(t, IconCellWidth, terminalDisplayWidthPlain(ansi.Strip(got)))
}

func TestTerminalDisplayWidthPlain_LightningTwoCells(t *testing.T) {
	assert.Equal(t, 2, terminalDisplayWidthPlain("⚡"))
}

func TestDisplayCellWidth_StyledBoxDrawingSeparatorMatchesAnsi(t *testing.T) {
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("│")
	assert.Equal(t, 1, ansi.StringWidth(sep))
	assert.Equal(t, 1, displayCellWidth(sep), "runewidth treats U+2502 as wide; TUI join uses ansi width 1")
}
