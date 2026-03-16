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
	slots         []slot.Slot
	filteredSlots []slot.Slot
	branches      []string
	cursor        int
	step          step
	input         textinput.Model
	filterInput   textinput.Model
	filterQuery   string
	result        Result
	noColor       bool
	err           error
}

func NewInteractiveModel(slots []slot.Slot, branches []string, noColor bool) Model {
	ti := textinput.New()
	ti.Placeholder = "branch name (prefix with + to create new)"
	ti.CharLimit = 256
	ti.Width = 50

	fi := textinput.New()
	fi.Placeholder = "type to filter..."
	fi.CharLimit = 128
	fi.Width = 30

	filtered := make([]slot.Slot, len(slots))
	copy(filtered, slots)

	m := Model{
		slots:         slots,
		filteredSlots: filtered,
		branches:      branches,
		step:          stepSlotSelect,
		input:         ti,
		filterInput:   fi,
		noColor:       noColor,
	}

	m.filterInput.Focus()

	return m
}

func (m Model) Init() tea.Cmd {
	return m.filterInput.Cursor.BlinkCmd()
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
	case "esc", "ctrl+c":
		m.step = stepAborted
		return m, tea.Quit
	case "up", "ctrl+k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "ctrl+j":
		if m.cursor < len(m.filteredSlots)-1 {
			m.cursor++
		}
		return m, nil
	case "enter":
		return m.selectCurrentSlot()
	}

	// Pass other keys to the filter input
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)

	newQuery := m.filterInput.Value()
	if newQuery != m.filterQuery {
		m.filterQuery = newQuery
		m.applyFilter()
	}

	return m, cmd
}

func (m *Model) applyFilter() {
	if m.filterQuery == "" {
		m.filteredSlots = make([]slot.Slot, len(m.slots))
		copy(m.filteredSlots, m.slots)
	} else {
		query := strings.ToLower(m.filterQuery)
		filtered := make([]slot.Slot, 0)
		for _, s := range m.slots {
			name := strings.ToLower(s.Name)
			branch := strings.ToLower(s.Branch)
			if strings.Contains(name, query) || strings.Contains(branch, query) {
				filtered = append(filtered, s)
			}
		}
		m.filteredSlots = filtered
	}
	if m.cursor >= len(m.filteredSlots) {
		m.cursor = max(0, len(m.filteredSlots)-1)
	}
}

func (m Model) selectCurrentSlot() (tea.Model, tea.Cmd) {
	if len(m.filteredSlots) == 0 {
		return m, nil
	}

	selected := m.filteredSlots[m.cursor]
	m.result.SlotName = selected.Name

	if selected.State == slot.SlotActive {
		m.step = stepDone
		return m, tea.Quit
	}

	m.step = stepBranchInput
	m.input.Focus()
	return m, m.input.Cursor.BlinkCmd()
}

func (m Model) updateBranchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.step = stepSlotSelect
		m.input.Reset()
		m.filterInput.Focus()
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
	prompt := "Select a slot:"
	if !m.noColor {
		prompt = StyleSelected.Render(prompt)
	}
	b.WriteString(prompt + "\n\n")

	b.WriteString(m.filterInput.View())
	b.WriteString("\n\n")

	for i, s := range m.filteredSlots {
		b.WriteString(renderSlotItem(s, i == m.cursor, m.noColor) + "\n")
	}

	b.WriteString("\n↑/↓ or ctrl+j/k: navigate  enter: select  esc: quit")
	return b.String()
}

func renderSlotItem(s slot.Slot, isSelected bool, noColor bool) string {
	cursor := "  "
	if isSelected {
		if noColor {
			cursor = "> "
		} else {
			cursor = StyleCursor.Render("> ")
		}
	}

	name := s.Name
	if !noColor {
		if isSelected {
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
	if !noColor {
		stateTag = StateStyle(state).Render(stateTag)
	}

	line := fmt.Sprintf("%s%s%s  %s", cursor, icon, name, stateTag)
	if s.State == slot.SlotActive {
		branch := s.Branch
		if !noColor {
			branch = StyleBranch.Render(branch)
		}
		line += "  " + branch
	}
	return line
}

func (m Model) viewBranchInput() string {
	var b strings.Builder
	selected := m.filteredSlots[m.cursor]

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
