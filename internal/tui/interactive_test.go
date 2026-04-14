package tui

import (
	"testing"

	"github.com/AquiTCD/git-slot/internal/slot"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSlots() []slot.Slot {
	return []slot.Slot{
		{Name: "wood", Icon: "🌱", State: slot.SlotEmpty},
		{Name: "fire", Icon: "🔥", State: slot.SlotActive, Branch: "feature/x", Path: "/slots/fire"},
		{Name: "earth", Icon: "🧱", State: slot.SlotEmpty},
	}
}

func newInteractiveTestModel(slots []slot.Slot, noColor bool) Model {
	return NewInteractiveModel(slots, nil, noColor)
}

func sendKey(m tea.Model, key string) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(Model)
}

func sendSpecialKey(m tea.Model, key tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(Model)
}

func TestNewInteractiveModel(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, stepSlotSelect, m.step)
	assert.Equal(t, filterFieldWidth(0), m.filterInput.Width)
}

func TestSlotSelect_MoveDown(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyDown)
	assert.Equal(t, 1, m2.cursor)
}

func TestSlotSelect_MoveUp(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyDown)
	m3 := sendSpecialKey(m2, tea.KeyUp)
	assert.Equal(t, 0, m3.cursor)
}

func TestSlotSelect_BoundsTop(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyUp)
	assert.Equal(t, 0, m2.cursor)
}

func TestSlotSelect_BoundsBottom(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyDown)
	m3 := sendSpecialKey(m2, tea.KeyDown)
	m4 := sendSpecialKey(m3, tea.KeyDown)
	assert.Equal(t, 2, m4.cursor)
}

func TestSlotSelect_EnterOnActive_DirectDone(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyDown)
	m3 := sendSpecialKey(m2, tea.KeyEnter)
	result, ok := m3.GetResult()
	require.True(t, ok)
	assert.Equal(t, "fire", result.SlotName)
	assert.Empty(t, result.BranchName)
}

func TestSlotSelect_EnterOnEmpty_GoesToBranchInput(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	assert.Equal(t, stepBranchInput, m2.step)
	assert.Equal(t, "wood", m2.result.SlotName)
}

func TestSlotSelect_Esc(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyEsc)
	assert.True(t, m2.Aborted())
}

func TestBranchInput_Esc_BackToSlotSelect(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	m3 := sendSpecialKey(m2, tea.KeyEsc)
	assert.Equal(t, stepSlotSelect, m3.step)
}

func TestView_SlotSelect(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	view := m.View()
	assert.Contains(t, view, "type to filter")
	assert.Contains(t, view, "3/3")
	assert.Contains(t, view, "wood")
	assert.Contains(t, view, "fire")
	assert.Contains(t, view, "earth")
	assert.Contains(t, view, "> ")
}

func TestView_SlotSelect_ActiveSlotShowsBranchLikeList(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyDown)
	view := m2.View()
	assert.Contains(t, view, "feature/x")
}

func TestView_BranchInput(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	view := m2.View()
	assert.Contains(t, view, "Slot:")
	assert.Contains(t, view, "wood")
	assert.Contains(t, view, "🌱")
	assert.Contains(t, view, "Enter branch name")
}

func TestView_WithColor(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), false)
	view := m.View()
	assert.Contains(t, view, "wood")
}

func TestRenderInteractiveSlotLine_SelectedPrefix(t *testing.T) {
	s := testSlots()[1]
	line := RenderInteractiveSlotLine(s, 4, true, true)
	assert.Contains(t, line, ">")
	assert.Contains(t, line, "fire")
	assert.Contains(t, line, "feature/x")
}

func TestRenderInteractiveSlotLine_ColorSelectedUsesBarNotGreater(t *testing.T) {
	s := testSlots()[1]
	line := RenderInteractiveSlotLine(s, 4, true, false)
	assert.NotContains(t, line, ">")
	assert.Contains(t, line, "fire")
}

func TestFilter_ShowsFilterInput(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	view := m.View()
	assert.Contains(t, view, "type to filter...")
	assert.Contains(t, view, "ctrl+j/k")
}

func TestFilter_FiltersByName(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendKey(m, "f")
	m3 := sendKey(m2, "i")
	m4 := sendKey(m3, "r")
	assert.Len(t, m4.filteredSlots, 1)
	assert.Equal(t, "fire", m4.filteredSlots[0].Name)
}

func TestFilter_FiltersByBranch(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	for _, ch := range "feature" {
		m = sendKey(m, string(ch))
	}
	assert.Len(t, m.filteredSlots, 1)
	assert.Equal(t, "fire", m.filteredSlots[0].Name)
}

func TestFilter_EmptyResult(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	for _, ch := range "zzz" {
		m = sendKey(m, string(ch))
	}
	assert.Empty(t, m.filteredSlots)
}

func TestFilter_EnterOnEmpty_Noop(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	for _, ch := range "zzz" {
		m = sendKey(m, string(ch))
	}
	m2 := sendSpecialKey(m, tea.KeyEnter)
	assert.Equal(t, stepSlotSelect, m2.step)
}

func TestFilter_CursorClamped(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m = sendSpecialKey(m, tea.KeyDown)
	m = sendSpecialKey(m, tea.KeyDown)
	assert.Equal(t, 2, m.cursor)
	m = sendKey(m, "f")
	m = sendKey(m, "i")
	m = sendKey(m, "r")
	assert.Equal(t, 0, m.cursor)
	assert.Len(t, m.filteredSlots, 1)
}

func TestCursorMove_NoBackgroundFetch(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Nil(t, cmd)
}

func TestWindowSizeMsg_SyncsFilterInputWidth(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2, ok := updated.(Model)
	require.True(t, ok)
	assert.Equal(t, 120, m2.width)
	assert.Equal(t, filterFieldWidth(120), m2.filterInput.Width)
}

func TestSanitizeTextInputFirstLine_DropsSecondLine(t *testing.T) {
	assert.Equal(t, "foo", sanitizeTextInputFirstLine("foo\nbar"))
	assert.Equal(t, "foo", sanitizeTextInputFirstLine("foo\r\nbar"))
}

func TestSanitizeTextInputFirstLine_CRWithoutLF(t *testing.T) {
	assert.Equal(t, "a b", sanitizeTextInputFirstLine("a\rb"))
}

func TestSanitizeTextInputFirstLine_FirstLineOnly(t *testing.T) {
	assert.Equal(t, "x", sanitizeTextInputFirstLine("x\ny"))
}
