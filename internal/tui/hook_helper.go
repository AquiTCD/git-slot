package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type HookAction int

const (
	ActionNone HookAction = iota
	ActionSymlink
	ActionCopy
)

func (a HookAction) String() string {
	switch a {
	case ActionSymlink:
		return "Link"
	case ActionCopy:
		return "Copy"
	default:
		return "None"
	}
}

func (a HookAction) Next() HookAction {
	return (a + 1) % 3
}

func (a HookAction) Prev() HookAction {
	return (a + 2) % 3
}

type HookItem struct {
	Path       string
	Action     HookAction
	IsDir      bool
	Expanded   bool
	Children   []*HookItem
	ChildCount int
}

type HookModel struct {
	items   []HookItem
	cursor  int
	done    bool
	aborted bool
	noColor bool
}

func NewHookModel(paths []string, noColor bool) HookModel {
	items := make([]HookItem, len(paths))
	for i, p := range paths {
		items[i] = HookItem{Path: p, Action: ActionNone}
	}
	return HookModel{
		items:   items,
		noColor: noColor,
	}
}

func NewHookModelFromItems(items []HookItem, noColor bool) HookModel {
	return HookModel{
		items:   items,
		noColor: noColor,
	}
}

func (m HookModel) Init() tea.Cmd {
	return nil
}

func (m HookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.aborted = true
			return m, tea.Quit
		case "up", "k", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j", "ctrl+j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "l":
			m.items[m.cursor].Action = ActionSymlink
		case "c":
			m.items[m.cursor].Action = ActionCopy
		case "n", "x":
			m.items[m.cursor].Action = ActionNone
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m HookModel) View() string {
	if m.aborted {
		return "Aborted.\n"
	}

	var b strings.Builder
	title := "Setup Post-Mount Hooks for Ignored Files"
	if !m.noColor {
		title = StyleSelected.Render(title)
	}
	b.WriteString(title + "\n\n")
	b.WriteString("Select action for each file (ctrl+j/k to move):\n")
	b.WriteString("(l)ink, (c)opy, (n)one\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			if m.noColor {
				cursor = "> "
			} else {
				cursor = StyleCursor.Render("> ")
			}
		}

		actionStr := fmt.Sprintf("[%s]", item.Action.String())
		if !m.noColor {
			switch item.Action {
			case ActionSymlink:
				actionStr = StyleSelected.Render(actionStr)
			case ActionCopy:
				actionStr = StyleBranch.Render(actionStr)
			}
		}

		b.WriteString(fmt.Sprintf("%s %-10s %s\n", cursor, actionStr, item.Path))
	}

	b.WriteString("\nenter: update configuration  q/esc: cancel")
	return b.String()
}

func (m HookModel) GetResults() []HookItem {
	return m.items
}

func (m HookModel) Aborted() bool {
	return m.aborted
}
