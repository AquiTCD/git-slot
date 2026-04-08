package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	_, _, _, leftCW, _ := slotSelectLayout(0)
	fi.Width = max(8, leftCW)

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
		_, _, _, leftCW, _ := slotSelectLayout(msg.Width)
		m.filterInput.Width = max(8, leftCW)
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
	innerWidth, leftWidth, rightWidth, leftContentW, rightContentW := slotSelectLayout(m.width)

	// Left and right columns must be built as one physical terminal row per logical UI row.
	// lipgloss Width + word wrap on either side adds extra lines and breaks index-pairing with
	// the │ separator (left wrap was the main cause of “missing” borders in the screenshot).
	leftBlock := m.buildLeftPane(leftContentW, leftWidth)
	rightBlock := m.buildRightPane(rightContentW, rightWidth)

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
	titleW := ansi.StringWidth(titleLine)
	if pad := innerWidth - titleW; pad > 0 {
		titleLine += strings.Repeat(" ", pad)
	}
	body := titleLine + "\n" + strings.Join(rows, "\n")

	// Do NOT set Width(innerWidth) here: lipgloss applies cellbuf.Wrap to the whole body at
	// (width - horizontal padding). Our grid rows are already innerWidth cells wide, which is
	// wider than that wrap limit, so every row gets hard-split and the │ layout shatters
	// (exactly the broken copy-paste the user saw).
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(0, 1).
		Render(body)
}

func (m Model) buildLeftPane(contentWidth int, leftColWidth int) string {
	if leftColWidth <= 0 {
		return ""
	}
	row := func(s string) string {
		return formatGridColumnLine(s, contentWidth, leftColWidth)
	}
	blank := strings.Repeat(" ", leftColWidth)

	prompt := render(StyleSelected, "Select a slot:", m.noColor)
	filterLine := strings.ReplaceAll(m.filterInput.View(), "\r\n", "\n")
	if i := strings.IndexByte(filterLine, '\n'); i >= 0 {
		filterLine = filterLine[:i]
	}

	lines := []string{
		row(prompt),
		blank,
		row(filterLine),
		blank,
	}
	for i, s := range m.filteredSlots {
		lines = append(lines, row(renderSlotItem(s, i == m.cursor, m.noColor)))
	}
	lines = append(lines, blank)
	help := "↑/↓ or ctrl+j/k: navigate  enter: select  esc: quit"
	lines = append(lines, row(help))
	return strings.Join(lines, "\n")
}

func (m Model) buildRightPane(contentWidth int, rightColWidth int) string {
	blank := strings.Repeat(" ", rightColWidth)
	row := func(s string) string {
		return formatGridColumnLine(s, contentWidth, rightColWidth)
	}

	if len(m.filteredSlots) == 0 {
		return row("(empty)")
	}
	s := m.filteredSlots[m.cursor]
	if s.State != slot.SlotActive {
		return row("(empty)")
	}

	icon := ""
	if s.Icon != "" {
		icon = s.Icon + " "
	}
	slotLabel := render(StyleSelected, icon+s.Name, m.noColor)
	branchLabel := render(StyleBranch, "  "+s.Branch, m.noColor)
	header := slotLabel + branchLabel

	if len(m.rightPane) == 0 {
		return strings.Join([]string{row(header), blank, blank, row("Loading...")}, "\n")
	}

	out := []string{row(header), blank, blank}
	for _, line := range m.rightPane {
		out = append(out, row(line))
	}
	return strings.Join(out, "\n")
}

// formatGridColumnLine is one physical row of a split-pane column: exactly colWidth
// terminal cells (one gutter space after │ + truncated text padded to contentWidth).
func formatGridColumnLine(line string, contentWidth int, colWidth int) string {
	if colWidth <= 0 {
		return ""
	}
	if contentWidth <= 0 {
		return strings.Repeat(" ", colWidth)
	}
	t := truncatePaneLine(line, contentWidth)
	s := " " + t
	w := ansi.StringWidth(s)
	if w < colWidth {
		s += strings.Repeat(" ", colWidth-w)
	}
	return s
}

// goldenRatio is φ (1.618…). Left column (slots + branches) is the larger segment:
// leftW : rightW ≈ φ : 1 over (inner − separator), i.e. left gets φ/(1+φ) ≈ 61.8% of that space.
const goldenRatio = 1.6180339887498948482045868364

// slotSelectLayout computes inner and column widths for the slot-select view. termWidth 0
// uses the same default as Bubble Tea before the first WindowSizeMsg (88 cols).
func slotSelectLayout(termWidth int) (innerW, leftW, rightW, leftContentW, rightContentW int) {
	if termWidth <= 0 {
		termWidth = 88
	}
	innerW = termWidth - 4 // RoundedBorder l+r + Padding l+r
	leftPad := 1
	splittable := innerW - 1 // room for │ between columns
	if splittable < 2 {
		splittable = 2
	}
	// Smaller column = right (log preview); larger = left (slot list).
	rightW = int(math.Round(float64(splittable) / (1 + goldenRatio)))
	leftW = splittable - rightW
	if rightW < 10 {
		rightW = 10
		leftW = splittable - rightW
	}
	if leftW < 1 {
		leftW = 1
		rightW = splittable - leftW
	}
	leftContentW = leftW - leftPad
	if leftContentW < 1 {
		leftContentW = 1
	}
	rightContentW = rightW - leftPad
	if rightContentW < 8 {
		rightContentW = 8
	}
	return
}

// truncatePaneLine limits one logical row to contentWidth terminal cells, preserving ANSI.
func truncatePaneLine(s string, maxCells int) string {
	if maxCells <= 0 {
		return ""
	}
	tail := "..."
	if ansi.StringWidth(tail) >= maxCells {
		return ansi.Truncate(s, maxCells, "")
	}
	return ansi.Truncate(s, maxCells, tail)
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
