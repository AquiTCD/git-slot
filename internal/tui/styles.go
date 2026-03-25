package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	renderer = lipgloss.DefaultRenderer()

	StyleIcon   lipgloss.Style
	StyleName   lipgloss.Style
	StyleBranch lipgloss.Style
	StyleHash   lipgloss.Style

	StyleStateEmpty  lipgloss.Style
	StyleStateActive lipgloss.Style
	StyleStateDirty  lipgloss.Style

	StyleDirtyMark lipgloss.Style
	StyleAheadMark lipgloss.Style

	StyleSelected lipgloss.Style
	StyleCursor   lipgloss.Style
)

func init() {
	// Force color profile before initializing styles if requested
	if os.Getenv("CLICOLOR_FORCE") != "" && os.Getenv("CLICOLOR_FORCE") != "0" {
		renderer.SetColorProfile(termenv.TrueColor)
	}

	StyleIcon = renderer.NewStyle().Foreground(lipgloss.Color("6"))
	StyleName = renderer.NewStyle().Bold(true)
	StyleBranch = renderer.NewStyle().Foreground(lipgloss.Color("4"))
	StyleHash = renderer.NewStyle().Foreground(lipgloss.Color("8"))

	StyleStateEmpty = renderer.NewStyle().Foreground(lipgloss.Color("8"))
	StyleStateActive = renderer.NewStyle().Foreground(lipgloss.Color("2"))
	StyleStateDirty = renderer.NewStyle().Foreground(lipgloss.Color("3"))

	StyleDirtyMark = renderer.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	StyleAheadMark = renderer.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)

	StyleSelected = renderer.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	StyleCursor = renderer.NewStyle().Foreground(lipgloss.Color("5"))
}

func render(s lipgloss.Style, text string, noColor bool) string {
	if noColor {
		return text
	}
	return s.Render(text)
}

func StateStyle(state string) lipgloss.Style {
	switch state {
	case "active":
		return StyleStateActive
	case "dirty":
		return StyleStateDirty
	default:
		return StyleStateEmpty
	}
}
