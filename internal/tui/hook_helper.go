package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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

func (item *HookItem) EffectiveAction() HookAction {
	if len(item.Children) == 0 {
		return item.Action
	}
	first := item.Children[0].Action
	for _, c := range item.Children[1:] {
		if c.Action != first {
			return HookAction(-1)
		}
	}
	return first
}

type ExpandFunc func(dir string) ([]*HookItem, error)

type HookModel struct {
	root        []*HookItem
	visible     []*HookItem
	cursor      int
	filterInput textinput.Model
	filterQuery string
	navStack    []string
	expandFn    ExpandFunc
	done        bool
	aborted     bool
	noColor     bool
}

func NewHookTreeModel(items []*HookItem, expandFn ExpandFunc, noColor bool) HookModel {
	fi := textinput.New()
	fi.Placeholder = "type to filter..."
	fi.CharLimit = 128
	fi.Width = 40
	fi.Focus()

	m := HookModel{
		root:        items,
		expandFn:    expandFn,
		noColor:     noColor,
		filterInput: fi,
	}
	m.rebuildVisible()
	return m
}

func (m *HookModel) currentItems() []*HookItem {
	if len(m.navStack) == 0 {
		return m.root
	}
	item := m.findItem(m.navStack[len(m.navStack)-1])
	if item != nil {
		return item.Children
	}
	return m.root
}

func (m *HookModel) findItem(path string) *HookItem {
	for _, item := range m.root {
		if item.Path == path {
			return item
		}
		for _, child := range item.Children {
			if child.Path == path {
				return child
			}
		}
	}
	return nil
}

func (m *HookModel) rebuildVisible() {
	items := m.currentItems()
	if m.filterQuery == "" {
		m.visible = items
	} else {
		query := strings.ToLower(m.filterQuery)
		var filtered []*HookItem
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Path), query) {
				filtered = append(filtered, item)
			}
		}
		m.visible = filtered
	}
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
}

func (m HookModel) Init() tea.Cmd {
	return m.filterInput.Cursor.BlinkCmd()
}

func (m HookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.Type {
	case tea.KeyCtrlC:
		m.aborted = true
		return m, tea.Quit

	case tea.KeyCtrlS:
		m.done = true
		return m, tea.Quit

	case tea.KeyUp, tea.KeyCtrlK:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyDown, tea.KeyCtrlJ:
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
		return m, nil

	case tea.KeyTab:
		if len(m.visible) > 0 {
			item := m.visible[m.cursor]
			newAction := item.Action.Next()
			item.Action = newAction
			for _, c := range item.Children {
				c.Action = newAction
			}
		}
		return m, nil

	case tea.KeyShiftTab:
		if len(m.visible) > 0 {
			item := m.visible[m.cursor]
			newAction := item.Action.Prev()
			item.Action = newAction
			for _, c := range item.Children {
				c.Action = newAction
			}
		}
		return m, nil

	case tea.KeyEnter, tea.KeyRight:
		if len(m.visible) > 0 {
			item := m.visible[m.cursor]
			if item.IsDir {
				if len(item.Children) == 0 && m.expandFn != nil {
					children, err := m.expandFn(item.Path)
					if err == nil && len(children) > 0 {
						for _, c := range children {
							c.Action = item.Action
						}
						item.Children = children
						item.ChildCount = len(children)
					}
				}
				if len(item.Children) > 0 {
					item.Expanded = true
					m.navStack = append(m.navStack, item.Path)
					m.filterInput.SetValue("")
					m.filterQuery = ""
					m.cursor = 0
					m.rebuildVisible()
				}
			}
		}
		return m, nil

	case tea.KeyLeft, tea.KeyCtrlH:
		if len(m.navStack) > 0 {
			m.navStack = m.navStack[:len(m.navStack)-1]
			m.filterInput.SetValue("")
			m.filterQuery = ""
			m.cursor = 0
			m.rebuildVisible()
		}
		return m, nil

	case tea.KeyEsc:
		if len(m.navStack) > 0 {
			m.navStack = m.navStack[:len(m.navStack)-1]
			m.filterInput.SetValue("")
			m.filterQuery = ""
			m.cursor = 0
			m.rebuildVisible()
		} else {
			m.aborted = true
			return m, tea.Quit
		}
		return m, nil

	case tea.KeyBackspace:
		if m.filterInput.Value() == "" && len(m.navStack) > 0 {
			m.navStack = m.navStack[:len(m.navStack)-1]
			m.filterQuery = ""
			m.cursor = 0
			m.rebuildVisible()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)

	newQuery := m.filterInput.Value()
	if newQuery != m.filterQuery {
		m.filterQuery = newQuery
		m.rebuildVisible()
	}

	return m, cmd
}

func (m HookModel) View() string {
	if m.aborted {
		return "Aborted.\n"
	}

	var b strings.Builder
	title := "Setup Post-Mount Hooks"
	if !m.noColor {
		title = StyleSelected.Render(title)
	}
	b.WriteString(title + "\n")

	if len(m.navStack) > 0 {
		dir := m.navStack[len(m.navStack)-1]
		fmt.Fprintf(&b, "▼ %s\n", dir)
	}

	b.WriteString(m.filterInput.View())
	b.WriteString("\n\n")

	for i, item := range m.visible {
		cursor := "  "
		if i == m.cursor {
			if m.noColor {
				cursor = "> "
			} else {
				cursor = StyleCursor.Render("> ")
			}
		}

		actionStr := m.formatAction(item)

		suffix := ""
		if item.IsDir && item.ChildCount > 0 && !item.Expanded {
			suffix = fmt.Sprintf(" (%d items)", item.ChildCount)
			if !m.noColor {
				suffix = StyleHash.Render(suffix)
			}
		}

		path := item.Path
		if !m.noColor && i == m.cursor {
			path = StyleSelected.Render(path)
		}

		fmt.Fprintf(&b, "%s%-10s %s%s\n", cursor, actionStr, path, suffix)
	}

	if len(m.navStack) > 0 {
		b.WriteString("\nctrl+j/k: move  tab: cycle mark  ←/esc: back  ctrl+s: confirm")
	} else {
		b.WriteString("\nctrl+j/k: move  tab: cycle mark  enter: expand  ctrl+s: confirm")
	}
	return b.String()
}

func (m HookModel) formatAction(item *HookItem) string {
	if item.IsDir && len(item.Children) > 0 {
		eff := item.EffectiveAction()
		if eff == HookAction(-1) {
			tag := "[Mixed]"
			if !m.noColor {
				tag = StyleStateDirty.Render(tag)
			}
			return tag
		}
		return m.styledAction(eff)
	}
	return m.styledAction(item.Action)
}

func (m HookModel) styledAction(a HookAction) string {
	tag := fmt.Sprintf("[%s]", a.String())
	if !m.noColor {
		switch a {
		case ActionSymlink:
			tag = StyleSelected.Render(tag)
		case ActionCopy:
			tag = StyleBranch.Render(tag)
		}
	}
	return tag
}

func (m HookModel) GetResults() []HookItem {
	var results []HookItem
	for _, item := range m.root {
		if item.Expanded && len(item.Children) > 0 {
			for _, child := range item.Children {
				results = append(results, *child)
			}
		} else {
			results = append(results, *item)
		}
	}
	return results
}

func (m HookModel) Aborted() bool {
	return m.aborted
}
