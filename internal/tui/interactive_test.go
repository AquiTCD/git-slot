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
		{Name: "fire", Icon: "🔥", State: slot.SlotActive, Branch: "feature/x"},
		{Name: "earth", Icon: "🧱", State: slot.SlotEmpty},
	}
}

func sendKey(m tea.Model, key string) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(Model)
}

func sendSpecialKey(m tea.Model, key tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(Model)
}

// --- Tests without filter (filter=false) ---

func TestNewInteractiveModel(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, stepSlotSelect, m.step)
	assert.False(t, m.filterEnabled)
}

func TestSlotSelect_MoveDown(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendKey(m, "j")
	assert.Equal(t, 1, m2.cursor)
}

func TestSlotSelect_MoveUp(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendKey(m, "j")
	m3 := sendKey(m2, "k")
	assert.Equal(t, 0, m3.cursor)
}

func TestSlotSelect_BoundsTop(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendKey(m, "k")
	assert.Equal(t, 0, m2.cursor)
}

func TestSlotSelect_BoundsBottom(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendKey(m, "j")
	m3 := sendKey(m2, "j")
	m4 := sendKey(m3, "j")
	assert.Equal(t, 2, m4.cursor)
}

func TestSlotSelect_EnterOnActive_DirectDone(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendKey(m, "j")
	m3 := sendSpecialKey(m2, tea.KeyEnter)
	result, ok := m3.GetResult()
	require.True(t, ok)
	assert.Equal(t, "fire", result.SlotName)
	assert.Empty(t, result.BranchName)
}

func TestSlotSelect_EnterOnEmpty_GoesToBranchInput(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	assert.Equal(t, stepBranchInput, m2.step)
	assert.Equal(t, "wood", m2.result.SlotName)
}

func TestSlotSelect_Quit(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendKey(m, "q")
	assert.True(t, m2.Aborted())
}

func TestSlotSelect_Esc(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendSpecialKey(m, tea.KeyEsc)
	assert.True(t, m2.Aborted())
}

func TestBranchInput_Esc_BackToSlotSelect(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	m3 := sendSpecialKey(m2, tea.KeyEsc)
	assert.Equal(t, stepSlotSelect, m3.step)
}

func TestView_SlotSelect(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	view := m.View()
	assert.Contains(t, view, "Select a slot:")
	assert.Contains(t, view, "wood")
	assert.Contains(t, view, "fire")
	assert.Contains(t, view, "earth")
	assert.Contains(t, view, "> ")
}

func TestView_BranchInput(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	view := m2.View()
	assert.Contains(t, view, "Slot: 🌱 wood")
	assert.Contains(t, view, "Enter branch name")
}

func TestView_WithColor(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, false, false)
	view := m.View()
	assert.Contains(t, view, "wood")
}

// --- Tests with filter (filter=true) ---

func TestFilter_Enabled_ShowsFilterInput(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, true)
	assert.True(t, m.filterEnabled)
	view := m.View()
	assert.Contains(t, view, "type to filter...")
	assert.Contains(t, view, "ctrl+j/k")
}

func TestFilter_Enabled_FiltersByName(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, true)
	// Type "fir" to filter
	m2 := sendKey(m, "f")
	m3 := sendKey(m2, "i")
	m4 := sendKey(m3, "r")
	assert.Len(t, m4.filteredSlots, 1)
	assert.Equal(t, "fire", m4.filteredSlots[0].Name)
}

func TestFilter_Enabled_FiltersByBranch(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, true)
	// Type "feature" to match branch
	for _, ch := range "feature" {
		m = sendKey(m, string(ch))
	}
	assert.Len(t, m.filteredSlots, 1)
	assert.Equal(t, "fire", m.filteredSlots[0].Name)
}

func TestFilter_Enabled_EmptyResult(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, true)
	for _, ch := range "zzz" {
		m = sendKey(m, string(ch))
	}
	assert.Empty(t, m.filteredSlots)
}

func TestFilter_Enabled_EnterOnEmpty_Noop(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, true)
	for _, ch := range "zzz" {
		m = sendKey(m, string(ch))
	}
	m2 := sendSpecialKey(m, tea.KeyEnter)
	// Should not crash or change state
	assert.Equal(t, stepSlotSelect, m2.step)
}

func TestFilter_Enabled_CursorClamped(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, true)
	// Move cursor to last item (index 2)
	m = sendSpecialKey(m, tea.KeyDown)
	m = sendSpecialKey(m, tea.KeyDown)
	assert.Equal(t, 2, m.cursor)
	// Now filter to 1 item — cursor should clamp
	m = sendKey(m, "f")
	m = sendKey(m, "i")
	m = sendKey(m, "r")
	assert.Equal(t, 0, m.cursor)
	assert.Len(t, m.filteredSlots, 1)
}

func TestFilter_Disabled_NoFilterView(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, false)
	view := m.View()
	assert.NotContains(t, view, "type to filter...")
	assert.Contains(t, view, "j/k")
}
