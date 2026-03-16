package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func testHookItems() []HookItem {
	return []HookItem{
		{Path: ".env", Action: ActionNone},
		{Path: "node_modules/", Action: ActionSymlink},
		{Path: "vendor/", Action: ActionCopy},
	}
}

func sendHookKey(m tea.Model, key string) HookModel {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(HookModel)
}

func sendHookSpecialKey(m tea.Model, key tea.KeyType) HookModel {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(HookModel)
}

func TestNewHookModel(t *testing.T) {
	m := NewHookModel([]string{".env", "vendor/"}, true)
	assert.Equal(t, 0, m.cursor)
	assert.Len(t, m.items, 2)
	assert.Equal(t, ActionNone, m.items[0].Action)
	assert.Equal(t, ActionNone, m.items[1].Action)
}

func TestNewHookModelFromItems(t *testing.T) {
	items := testHookItems()
	m := NewHookModelFromItems(items, true)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, ActionNone, m.items[0].Action)
	assert.Equal(t, ActionSymlink, m.items[1].Action)
	assert.Equal(t, ActionCopy, m.items[2].Action)
}

func TestHookModel_NavigateDown(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookSpecialKey(m, tea.KeyDown)
	assert.Equal(t, 1, m2.cursor)
}

func TestHookModel_NavigateUp(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookSpecialKey(m, tea.KeyDown)
	m3 := sendHookSpecialKey(m2, tea.KeyUp)
	assert.Equal(t, 0, m3.cursor)
}

func TestHookModel_BoundsTop(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookSpecialKey(m, tea.KeyUp)
	assert.Equal(t, 0, m2.cursor)
}

func TestHookModel_BoundsBottom(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookSpecialKey(m, tea.KeyDown)
	m3 := sendHookSpecialKey(m2, tea.KeyDown)
	m4 := sendHookSpecialKey(m3, tea.KeyDown)
	assert.Equal(t, 2, m4.cursor)
}

func TestHookModel_SetLink(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookKey(m, "l")
	assert.Equal(t, ActionSymlink, m2.items[0].Action)
}

func TestHookModel_SetCopy(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookKey(m, "c")
	assert.Equal(t, ActionCopy, m2.items[0].Action)
}

func TestHookModel_SetNone(t *testing.T) {
	items := testHookItems()
	items[0].Action = ActionSymlink
	m := NewHookModelFromItems(items, true)
	m2 := sendHookKey(m, "n")
	assert.Equal(t, ActionNone, m2.items[0].Action)
}

func TestHookModel_SetNoneWithX(t *testing.T) {
	items := testHookItems()
	items[0].Action = ActionCopy
	m := NewHookModelFromItems(items, true)
	m2 := sendHookKey(m, "x")
	assert.Equal(t, ActionNone, m2.items[0].Action)
}

func TestHookModel_Enter_Done(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookSpecialKey(m, tea.KeyEnter)
	assert.True(t, m2.done)
	assert.False(t, m2.Aborted())
}

func TestHookModel_Escape_Aborted(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookKey(m, "q")
	assert.True(t, m2.Aborted())
}

func TestHookModel_GetResults(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookKey(m, "l")
	results := m2.GetResults()
	assert.Len(t, results, 3)
	assert.Equal(t, ActionSymlink, results[0].Action)
}

func TestHookModel_View_NoColor(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	view := m.View()
	assert.Contains(t, view, "Setup Post-Mount Hooks")
	assert.Contains(t, view, ".env")
	assert.Contains(t, view, "node_modules/")
	assert.Contains(t, view, "> ")
}

func TestHookModel_View_Aborted(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookKey(m, "q")
	assert.Equal(t, "Aborted.\n", m2.View())
}

func TestHookModel_NavigateWithJK(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	m2 := sendHookKey(m, "j")
	assert.Equal(t, 1, m2.cursor)
	m3 := sendHookKey(m2, "k")
	assert.Equal(t, 0, m3.cursor)
}

func TestHookAction_String(t *testing.T) {
	assert.Equal(t, "None", ActionNone.String())
	assert.Equal(t, "Link", ActionSymlink.String())
	assert.Equal(t, "Copy", ActionCopy.String())
}

func TestHookModel_Init_ReturnsNil(t *testing.T) {
	m := NewHookModelFromItems(testHookItems(), true)
	assert.Nil(t, m.Init())
}
