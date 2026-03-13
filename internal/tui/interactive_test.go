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

func TestNewInteractiveModel(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, stepSlotSelect, m.step)
}

func TestSlotSelect_MoveDown(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendKey(m, "j")
	assert.Equal(t, 1, m2.cursor)
}

func TestSlotSelect_MoveUp(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendKey(m, "j")
	m3 := sendKey(m2, "k")
	assert.Equal(t, 0, m3.cursor)
}

func TestSlotSelect_BoundsTop(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendKey(m, "k")
	assert.Equal(t, 0, m2.cursor)
}

func TestSlotSelect_BoundsBottom(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendKey(m, "j")
	m3 := sendKey(m2, "j")
	m4 := sendKey(m3, "j")
	assert.Equal(t, 2, m4.cursor)
}

func TestSlotSelect_EnterOnActive_DirectDone(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendKey(m, "j")
	m3 := sendSpecialKey(m2, tea.KeyEnter)
	result, ok := m3.GetResult()
	require.True(t, ok)
	assert.Equal(t, "fire", result.SlotName)
	assert.Empty(t, result.BranchName)
}

func TestSlotSelect_EnterOnEmpty_GoesToBranchInput(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	assert.Equal(t, stepBranchInput, m2.step)
	assert.Equal(t, "wood", m2.result.SlotName)
}

func TestSlotSelect_Quit(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendKey(m, "q")
	assert.True(t, m2.Aborted())
}

func TestSlotSelect_Esc(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendSpecialKey(m, tea.KeyEsc)
	assert.True(t, m2.Aborted())
}

func TestBranchInput_Esc_BackToSlotSelect(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	m3 := sendSpecialKey(m2, tea.KeyEsc)
	assert.Equal(t, stepSlotSelect, m3.step)
}

func TestView_SlotSelect(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	view := m.View()
	assert.Contains(t, view, "Select a slot:")
	assert.Contains(t, view, "wood")
	assert.Contains(t, view, "fire")
	assert.Contains(t, view, "earth")
	assert.Contains(t, view, "> ")
}

func TestView_BranchInput(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	view := m2.View()
	assert.Contains(t, view, "Slot: 🌱 wood")
	assert.Contains(t, view, "Enter branch name")
}

func TestView_WithColor(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, false)
	view := m.View()
	assert.Contains(t, view, "wood")
}
