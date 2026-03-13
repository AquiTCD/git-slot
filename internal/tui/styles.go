package tui

import "github.com/charmbracelet/lipgloss"

var (
	StyleIcon   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	StyleName   = lipgloss.NewStyle().Bold(true)
	StyleBranch = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	StyleHash   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	StyleStateEmpty  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	StyleStateActive = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	StyleStateDirty  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	StyleDirtyMark = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)

	StyleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	StyleCursor   = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
)

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
