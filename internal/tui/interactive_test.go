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

var noopFetcher = func(path string, n int, format string) ([]string, error) {
	return nil, nil
}

func newInteractiveTestModel(slots []slot.Slot, noColor bool) Model {
	return NewInteractiveModel(slots, nil, noColor, noopFetcher, 5, "%h %s")
}

func sendKey(m tea.Model, key string) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(Model)
}

func sendSpecialKey(m tea.Model, key tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(Model)
}

// --- Navigation Tests ---

func TestNewInteractiveModel(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, stepSlotSelect, m.step)
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
	assert.Contains(t, view, "Select a slot:")
	assert.Contains(t, view, "wood")
	assert.Contains(t, view, "fire")
	assert.Contains(t, view, "earth")
	assert.Contains(t, view, "> ")
}

func TestView_BranchInput(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m2 := sendSpecialKey(m, tea.KeyEnter)
	view := m2.View()
	assert.Contains(t, view, "Slot: 🌱 wood")
	assert.Contains(t, view, "Enter branch name")
}

func TestView_WithColor(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), false)
	view := m.View()
	assert.Contains(t, view, "wood")
}

// --- Tests with filter ---

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

// --- Right pane / log preview tests ---

// ENH-P2-T1: empty slot => right pane shows "(empty)"
func TestRightPane_EmptySlot_ShowsEmptyLabel(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	// cursor starts at 0 = wood (empty)
	rp := m.buildRightPane()
	assert.Equal(t, "(empty)", rp)
}

// ENH-P2-T2: active slot with loaded logs => right pane shows those lines
func TestRightPane_ActiveSlot_ShowsLogs(t *testing.T) {
	logs := []string{"abc1234 feat: add auth", "def5678 fix: null pointer"}
	fetcher := func(path string, n int, format string) ([]string, error) {
		return logs, nil
	}
	m := NewInteractiveModel(testSlots(), nil, true, fetcher, 5, "%h %s")
	m.cursor = 1 // fire (active)
	m.rightPane = logs
	rp := m.buildRightPane()
	assert.Contains(t, rp, "abc1234 feat: add auth")
	assert.Contains(t, rp, "def5678 fix: null pointer")
}

// ENH-P2-T3: logLoadedMsg updates rightPane when slotPath matches current cursor
func TestLogLoadedMsg_UpdatesRightPane(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m.cursor = 1 // fire, Path="/slots/fire"
	updated, _ := m.Update(logLoadedMsg{
		slotPath: "/slots/fire",
		lines:    []string{"abc feat: x", "def fix: y"},
	})
	m2 := updated.(Model)
	assert.Equal(t, []string{"abc feat: x", "def fix: y"}, m2.rightPane)
}

// ENH-P2-T3b: stale logLoadedMsg (wrong slotPath) does NOT update rightPane
func TestLogLoadedMsg_StaleMsg_DoesNotUpdateRightPane(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m.cursor = 1 // fire, Path="/slots/fire"
	m.rightPane = []string{"existing log"}
	updated, _ := m.Update(logLoadedMsg{
		slotPath: "/slots/wood", // stale: cursor has moved away
		lines:    []string{"stale log"},
	})
	m2 := updated.(Model)
	assert.Equal(t, []string{"existing log"}, m2.rightPane, "stale msg should not overwrite rightPane")
}

// ENH-P2-T4: active slot with no logs yet => "Loading..."
func TestRightPane_ActiveSlot_LoadingState(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m.cursor = 1 // fire (active), rightPane is nil
	rp := m.buildRightPane()
	assert.Equal(t, "Loading...", rp)
}

// ENH-P2-T5: cursor move triggers non-nil log fetch cmd
func TestCursorMove_TriggersLogFetch(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	// cursor moves to 1 (fire = active) => fetchLogsForCurrent should return non-nil cmd
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.NotNil(t, cmd, "moving cursor to active slot should return a log-fetch cmd")
}

// ENH-P2-T6: no fetcher => fetchLogsForCurrent returns nil
func TestFetchLogsForCurrent_NoFetcher_ReturnsNil(t *testing.T) {
	m := NewInteractiveModel(testSlots(), nil, true, nil, 5, "%h %s")
	cmd := m.fetchLogsForCurrent()
	assert.Nil(t, cmd)
}

// ENH-P2-T7: view contains right pane content in split layout
func TestView_SplitLayout_ContainsRightPane(t *testing.T) {
	m := newInteractiveTestModel(testSlots(), true)
	m.cursor = 1
	m.rightPane = []string{"abc feat: add x"}
	view := m.View()
	assert.Contains(t, view, "abc feat: add x")
	assert.Contains(t, view, "Select a slot:")
}
