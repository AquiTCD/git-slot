package tui

import (
	"fmt"
	"strings"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type logLoadedMsg struct {
	slotPath string
	lines    []string
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
	logFetcher    func(path string, n int, format string) ([]string, error)
	logLines      int
	logFormat     string
	rightPane     []string
	width         int
	height        int
}

func NewInteractiveModel(
	slots []slot.Slot,
	branches []string,
	noColor bool,
	logFetcher func(path string, n int, format string) ([]string, error),
	logLines int,
	logFormat string,
) Model {
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
		logFetcher:    logFetcher,
		logLines:      logLines,
		logFormat:     logFormat,
	}

	m.filterInput.Focus()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.filterInput.Cursor.BlinkCmd(), m.fetchLogsForCurrent())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case logLoadedMsg:
		if len(m.filteredSlots) > 0 && m.filteredSlots[m.cursor].Path == msg.slotPath {
			m.rightPane = msg.lines
		}
		return m, nil
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
			return m, m.fetchLogsForCurrent()
		}
		return m, nil
	case "down", "ctrl+j":
		if m.cursor < len(m.filteredSlots)-1 {
			m.cursor++
			return m, m.fetchLogsForCurrent()
		}
		return m, nil
	case "enter":
		return m.selectCurrentSlot()
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)

	newQuery := m.filterInput.Value()
	if newQuery != m.filterQuery {
		m.filterQuery = newQuery
		m.applyFilter()
		return m, tea.Batch(cmd, m.fetchLogsForCurrent())
	}

	return m, cmd
}

func (m Model) fetchLogsForCurrent() tea.Cmd {
	if len(m.filteredSlots) == 0 || m.logFetcher == nil {
		return nil
	}
	s := m.filteredSlots[m.cursor]
	if s.State != slot.SlotActive {
		return func() tea.Msg {
			return logLoadedMsg{slotPath: s.Path, lines: nil}
		}
	}
	fetcher := m.logFetcher
	n := m.logLines
	format := m.logFormat
	path := s.Path
	return func() tea.Msg {
		lines, _ := fetcher(path, n, format)
		return logLoadedMsg{slotPath: path, lines: lines}
	}
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
	totalWidth := m.width
	if totalWidth == 0 {
		totalWidth = 88
	}

	// Outer: RoundedBorder l+r (2) + Padding l+r (2) = 4 overhead
	innerWidth := totalWidth - 4
	// Left pane gets 1 char left padding; │ separator = 1 char
	leftPad := 1
	leftWidth := innerWidth * 2 / 5
	leftContentWidth := leftWidth - leftPad
	rightWidth := innerWidth - leftWidth - 1
	if rightWidth < 10 {
		rightWidth = 10
	}

	// Render each pane as a fixed-width block.
	// lipgloss guarantees every output line is padded to exactly the specified width,
	// and long lines wrap within the pane rather than overflowing into the other pane.
	leftBlock := lipgloss.NewStyle().Width(leftContentWidth).PaddingLeft(leftPad).Render(m.buildLeftPane())
	rightBlock := lipgloss.NewStyle().Width(rightWidth).Render(m.buildRightPane(rightWidth))

	leftLines := strings.Split(leftBlock, "\n")
	rightLines := strings.Split(rightBlock, "\n")

	// Pad both sides to equal height so │ runs the full length
	height := max(len(leftLines), len(rightLines))
	emptyLeft := strings.Repeat(" ", leftWidth)
	emptyRight := strings.Repeat(" ", rightWidth)
	for len(leftLines) < height {
		leftLines = append(leftLines, emptyLeft)
	}
	for len(rightLines) < height {
		rightLines = append(rightLines, emptyRight)
	}

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("│")

	rows := make([]string, height)
	for i := range leftLines {
		rows[i] = leftLines[i] + sep + rightLines[i]
	}

	// "git slot" title as first content line inside the box
	titleLine := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render("git slot")
	body := titleLine + "\n" + strings.Join(rows, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(0, 1).
		Width(innerWidth).
		Render(body)
}

func (m Model) buildLeftPane() string {
	var b strings.Builder
	prompt := render(StyleSelected, "Select a slot:", m.noColor)
	b.WriteString(prompt + "\n\n")

	b.WriteString(m.filterInput.View())
	b.WriteString("\n\n")

	for i, s := range m.filteredSlots {
		b.WriteString(renderSlotItem(s, i == m.cursor, m.noColor) + "\n")
	}

	b.WriteString("\n↑/↓ or ctrl+j/k: navigate  enter: select  esc: quit")
	return b.String()
}

func (m Model) buildRightPane(_ int) string {
	if len(m.filteredSlots) == 0 {
		return "(empty)"
	}
	s := m.filteredSlots[m.cursor]
	if s.State != slot.SlotActive {
		return "(empty)"
	}

	icon := ""
	if s.Icon != "" {
		icon = s.Icon + " "
	}
	slotLabel := render(StyleSelected, icon+s.Name, m.noColor)
	branchLabel := render(StyleBranch, "  "+s.Branch, m.noColor)
	title := slotLabel + branchLabel

	if len(m.rightPane) == 0 {
		return title + "\n\nLoading..."
	}
	return title + "\n\n" + strings.Join(m.rightPane, "\n")
}

func renderSlotItem(s slot.Slot, isSelected bool, noColor bool) string {
	cursor := "  "
	if isSelected {
		cursor = render(StyleCursor, "> ", noColor)
	}

	name := s.Name
	if isSelected {
		name = render(StyleSelected, name, noColor)
	} else {
		name = render(StyleName, name, noColor)
	}

	icon := ""
	if s.Icon != "" {
		icon = s.Icon + " "
	}

	state := s.DisplayState()
	stateTag := render(StateStyle(state), fmt.Sprintf("[%s]", state), noColor)

	line := fmt.Sprintf("%s%s%s  %s", cursor, icon, name, stateTag)
	if s.State == slot.SlotActive {
		line += "  " + render(StyleBranch, s.Branch, noColor)
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

	fmt.Fprintf(&b, "Slot: %s\n\n", name)
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
