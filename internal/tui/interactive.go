package tui

import (
	"fmt"
	"strings"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type step int

const (
	stepSlotSelect step = iota
	stepBranchInput
	stepDone
	stepAborted
)

type Result struct {
	SlotName     string
	BranchName   string
	CreateBranch bool
}

type Model struct {
	slots    []slot.Slot
	branches []string
	cursor   int
	step     step
	input    textinput.Model
	result   Result
	noColor  bool
	err      error
}

func NewInteractiveModel(slots []slot.Slot, branches []string, noColor bool) Model {
	ti := textinput.New()
	ti.Placeholder = "branch name (prefix with + to create new)"
	ti.CharLimit = 256
	ti.Width = 50

	return Model{
		slots:    slots,
		branches: branches,
		step:     stepSlotSelect,
		input:    ti,
		noColor:  noColor,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case stepSlotSelect:
			return m.updateSlotSelect(msg)
		case stepBranchInput:
			return m.updateBranchInput(msg)
		}
	}
	return m, nil
}

func (m Model) updateSlotSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.step = stepAborted
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.slots)-1 {
			m.cursor++
		}
	case "enter":
		selected := m.slots[m.cursor]
		m.result.SlotName = selected.Name

		if selected.State == slot.SlotActive {
			m.step = stepDone
			return m, tea.Quit
		}

		m.step = stepBranchInput
		m.input.Focus()
		return m, m.input.Cursor.BlinkCmd()
	}
	return m, nil
}

func (m Model) updateBranchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.step = stepSlotSelect
		m.input.Reset()
		return m, nil
	case "ctrl+c":
		m.step = stepAborted
		return m, tea.Quit
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			return m, nil
		}
		if strings.HasPrefix(val, "+") {
			m.result.CreateBranch = true
			m.result.BranchName = strings.TrimPrefix(val, "+")
		} else {
			m.result.BranchName = val
		}
		m.step = stepDone
		return m, tea.Quit
	case "tab":
		return m.completeBranch()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) completeBranch() (tea.Model, tea.Cmd) {
	prefix := m.input.Value()
	if prefix == "" {
		return m, nil
	}
	for _, b := range m.branches {
		if strings.HasPrefix(b, prefix) {
			m.input.SetValue(b)
			m.input.SetCursor(len(b))
			break
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.step {
	case stepSlotSelect:
		return m.viewSlotSelect()
	case stepBranchInput:
		return m.viewBranchInput()
	default:
		return ""
	}
}

func (m Model) viewSlotSelect() string {
	var b strings.Builder
	b.WriteString("Select a slot:\n\n")

	for i, s := range m.slots {
		cursor := "  "
		if i == m.cursor {
			if m.noColor {
				cursor = "> "
			} else {
				cursor = StyleCursor.Render("> ")
			}
		}

		name := s.Name
		if !m.noColor {
			if i == m.cursor {
				name = StyleSelected.Render(name)
			} else {
				name = StyleName.Render(name)
			}
		}

		icon := ""
		if s.Icon != "" {
			icon = s.Icon + " "
		}

		state := s.DisplayState()
		stateTag := fmt.Sprintf("[%s]", state)
		if !m.noColor {
			stateTag = StateStyle(state).Render(stateTag)
		}

		line := fmt.Sprintf("%s%s%s  %s", cursor, icon, name, stateTag)
		if s.State == slot.SlotActive {
			branch := s.Branch
			if !m.noColor {
				branch = StyleBranch.Render(branch)
			}
			line += "  " + branch
		}

		b.WriteString(line + "\n")
	}

	b.WriteString("\n↑/↓ or j/k: navigate  enter: select  q/esc: quit")
	return b.String()
}

func (m Model) viewBranchInput() string {
	var b strings.Builder
	selected := m.slots[m.cursor]

	name := selected.Name
	if selected.Icon != "" {
		name = selected.Icon + " " + name
	}

	b.WriteString(fmt.Sprintf("Slot: %s\n\n", name))
	b.WriteString("Enter branch name (prefix with + to create new):\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n\ntab: complete  enter: confirm  esc: back")
	return b.String()
}

func (m Model) GetResult() (Result, bool) {
	if m.step == stepDone {
		return m.result, true
	}
	return Result{}, false
}

func (m Model) Aborted() bool {
	return m.step == stepAborted
}
