package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTreeItems() []*HookItem {
	return []*HookItem{
		{Path: ".env", Action: ActionNone},
		{Path: "node_modules/", IsDir: true},
		{Path: "vendor/", IsDir: true, ChildCount: 3, Children: []*HookItem{
			{Path: "vendor/a.txt"},
			{Path: "vendor/b.txt"},
			{Path: "vendor/c.txt"},
		}},
	}
}

func noopExpand(dir string) ([]*HookItem, error) {
	return nil, nil
}

func newTestModel(items []*HookItem) HookModel {
	return NewHookTreeModel(items, noopExpand, true)
}

func hookSendKey(m tea.Model, key string) HookModel {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(HookModel)
}

func hookSendSpecial(m tea.Model, key tea.KeyType) HookModel {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(HookModel)
}

// --- Navigation ---

func TestHookTreeModel_NavigateDown(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyDown)
	assert.Equal(t, 1, m2.cursor)
}

func TestHookTreeModel_NavigateUp(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyDown)
	m3 := hookSendSpecial(m2, tea.KeyUp)
	assert.Equal(t, 0, m3.cursor)
}

func TestHookTreeModel_CtrlJK(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyCtrlJ)
	assert.Equal(t, 1, m2.cursor)
	m3 := hookSendSpecial(m2, tea.KeyCtrlK)
	assert.Equal(t, 0, m3.cursor)
}

func TestHookTreeModel_BoundsTop(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyUp)
	assert.Equal(t, 0, m2.cursor)
}

func TestHookTreeModel_BoundsBottom(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyDown)
	m3 := hookSendSpecial(m2, tea.KeyDown)
	m4 := hookSendSpecial(m3, tea.KeyDown)
	assert.Equal(t, 2, m4.cursor)
}

// --- Tab toggle ---

func TestHookTreeModel_TabCyclesAction(t *testing.T) {
	m := newTestModel(testTreeItems())
	assert.Equal(t, ActionNone, m.visible[0].Action)

	m2 := hookSendSpecial(m, tea.KeyTab)
	assert.Equal(t, ActionSymlink, m2.visible[0].Action)

	m3 := hookSendSpecial(m2, tea.KeyTab)
	assert.Equal(t, ActionCopy, m3.visible[0].Action)

	m4 := hookSendSpecial(m3, tea.KeyTab)
	assert.Equal(t, ActionNone, m4.visible[0].Action)
}

func TestHookTreeModel_ShiftTabCyclesReverse(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyShiftTab)
	assert.Equal(t, ActionCopy, m2.visible[0].Action)
}

func TestHookTreeModel_TabOnAggregatedDir_AppliesAll(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)
	assert.Equal(t, 2, m.cursor)

	m = hookSendSpecial(m, tea.KeyTab)
	assert.Equal(t, ActionSymlink, m.visible[2].Action)
	for _, c := range m.visible[2].Children {
		assert.Equal(t, ActionSymlink, c.Action, "child %s should be Link", c.Path)
	}
}

// --- Expand / Collapse ---

func TestHookTreeModel_EnterExpandsDir(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)

	m = hookSendSpecial(m, tea.KeyEnter)
	assert.Len(t, m.navStack, 1)
	assert.Equal(t, "vendor/", m.navStack[0])
	require.Len(t, m.visible, 3)
	assert.Equal(t, "vendor/a.txt", m.visible[0].Path)
}

func TestHookTreeModel_RightExpandsDir(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)

	m = hookSendSpecial(m, tea.KeyRight)
	assert.Len(t, m.navStack, 1)
}

func TestHookTreeModel_LeftCollapses(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyEnter)
	assert.Len(t, m.navStack, 1)

	m = hookSendSpecial(m, tea.KeyLeft)
	assert.Len(t, m.navStack, 0)
	assert.Len(t, m.visible, 3)
}

func TestHookTreeModel_EscCollapses(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyEnter)

	m = hookSendSpecial(m, tea.KeyEsc)
	assert.Len(t, m.navStack, 0)
	assert.False(t, m.aborted)
}

func TestHookTreeModel_EscAtTopAborts(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyEsc)
	assert.True(t, m.aborted)
}

func TestHookTreeModel_EnterOnFile_NoOp(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyEnter)
	assert.Len(t, m.navStack, 0)
}

func TestHookTreeModel_OnDemandExpand(t *testing.T) {
	items := []*HookItem{
		{Path: "secret/", IsDir: true},
	}
	expandCalled := false
	expand := func(dir string) ([]*HookItem, error) {
		expandCalled = true
		assert.Equal(t, "secret/", dir)
		return []*HookItem{
			{Path: "secret/key.pem"},
			{Path: "secret/cert.pem"},
		}, nil
	}

	m := NewHookTreeModel(items, expand, true)
	m = hookSendSpecial(m, tea.KeyEnter)

	assert.True(t, expandCalled)
	assert.Len(t, m.navStack, 1)
	require.Len(t, m.visible, 2)
	assert.Equal(t, "secret/key.pem", m.visible[0].Path)
}

// --- Filter ---

func TestHookTreeModel_FilterReducesVisible(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendKey(m, "e")
	m = hookSendKey(m, "n")
	m = hookSendKey(m, "v")

	assert.Equal(t, ".env", m.visible[0].Path)
}

func TestHookTreeModel_FilterInExpandedDir(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyEnter)

	m = hookSendKey(m, "a")
	require.Len(t, m.visible, 1)
	assert.Equal(t, "vendor/a.txt", m.visible[0].Path)
}

func TestHookTreeModel_BackspaceEmptyCollapses(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyEnter)
	assert.Len(t, m.navStack, 1)

	m = hookSendSpecial(m, tea.KeyBackspace)
	assert.Len(t, m.navStack, 0)
}

// --- Confirm / Abort ---

func TestHookTreeModel_CtrlSConfirms(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyCtrlS)
	assert.True(t, m2.done)
	assert.False(t, m2.aborted)
}

func TestHookTreeModel_CtrlCAborts(t *testing.T) {
	m := newTestModel(testTreeItems())
	m2 := hookSendSpecial(m, tea.KeyCtrlC)
	assert.True(t, m2.aborted)
}

// --- GetResults ---

func TestHookTreeModel_GetResults_Flat(t *testing.T) {
	m := newTestModel(testTreeItems())
	results := m.GetResults()
	assert.Len(t, results, 3)
	assert.Equal(t, ".env", results[0].Path)
	assert.Equal(t, "node_modules/", results[1].Path)
	assert.Equal(t, "vendor/", results[2].Path)
}

func TestHookTreeModel_GetResults_ExpandedReturnsChildren(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyEnter)

	m = hookSendSpecial(m, tea.KeyLeft)

	results := m.GetResults()
	assert.Len(t, results, 5)
	assert.Equal(t, ".env", results[0].Path)
	assert.Equal(t, "node_modules/", results[1].Path)
	assert.Equal(t, "vendor/a.txt", results[2].Path)
}

// --- View ---

func TestHookTreeModel_View_NoColor(t *testing.T) {
	m := newTestModel(testTreeItems())
	view := m.View()
	assert.Contains(t, view, "Setup Post-Mount Hooks")
	assert.Contains(t, view, ".env")
	assert.Contains(t, view, "vendor/")
	assert.Contains(t, view, "(3 items)")
	assert.Contains(t, view, "> ")
}

func TestHookTreeModel_View_Aborted(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyEsc)
	assert.Equal(t, "Aborted.\n", m.View())
}

func TestHookTreeModel_View_ExpandedShowsBreadcrumb(t *testing.T) {
	m := newTestModel(testTreeItems())
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyDown)
	m = hookSendSpecial(m, tea.KeyEnter)
	view := m.View()
	assert.Contains(t, view, "▼ vendor/")
}

// --- EffectiveAction ---

func TestEffectiveAction_Uniform(t *testing.T) {
	item := &HookItem{
		Children: []*HookItem{
			{Action: ActionSymlink},
			{Action: ActionSymlink},
		},
	}
	assert.Equal(t, ActionSymlink, item.EffectiveAction())
}

func TestEffectiveAction_Mixed(t *testing.T) {
	item := &HookItem{
		Children: []*HookItem{
			{Action: ActionSymlink},
			{Action: ActionCopy},
		},
	}
	assert.Equal(t, HookAction(-1), item.EffectiveAction())
}

func TestEffectiveAction_NoChildren(t *testing.T) {
	item := &HookItem{Action: ActionCopy}
	assert.Equal(t, ActionCopy, item.EffectiveAction())
}

// --- HookAction helpers ---

func TestHookAction_String(t *testing.T) {
	assert.Equal(t, "None", ActionNone.String())
	assert.Equal(t, "Link", ActionSymlink.String())
	assert.Equal(t, "Copy", ActionCopy.String())
}

func TestHookAction_Next(t *testing.T) {
	assert.Equal(t, ActionSymlink, ActionNone.Next())
	assert.Equal(t, ActionCopy, ActionSymlink.Next())
	assert.Equal(t, ActionNone, ActionCopy.Next())
}

func TestHookAction_Prev(t *testing.T) {
	assert.Equal(t, ActionCopy, ActionNone.Prev())
	assert.Equal(t, ActionNone, ActionSymlink.Prev())
	assert.Equal(t, ActionSymlink, ActionCopy.Prev())
}

func TestHookTreeModel_Init_ReturnsBlink(t *testing.T) {
	m := newTestModel(testTreeItems())
	assert.NotNil(t, m.Init())
}
