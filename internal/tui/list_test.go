package tui

import (
	"strings"
	"testing"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/stretchr/testify/assert"
)

func TestRenderSlotList_Empty(t *testing.T) {
	result := RenderSlotList(nil, true)
	assert.Equal(t, "No slots defined.", result)
}

func TestRenderSlotList_NoColor_AllEmpty(t *testing.T) {
	slots := []slot.Slot{
		{Name: "wood", Icon: "🌱"},
		{Name: "fire", Icon: "🔥"},
	}
	result := RenderSlotList(slots, true)
	lines := strings.Split(result, "\n")
	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "wood")
	assert.Contains(t, lines[0], "[empty]")
	assert.Contains(t, lines[0], "🌱")
	assert.Contains(t, lines[1], "fire")
	assert.Contains(t, lines[1], "🔥")
}

func TestRenderSlotList_NoColor_ActiveSlot(t *testing.T) {
	slots := []slot.Slot{
		{Name: "wood", Icon: "🌱", State: slot.SlotActive, Branch: "feature/x", HeadHash: "abc1234"},
	}
	result := RenderSlotList(slots, true)
	assert.Contains(t, result, "[active]")
	assert.Contains(t, result, "feature/x")
	assert.Contains(t, result, "(abc1234)")
}

func TestRenderSlotList_NoColor_DirtySlot(t *testing.T) {
	slots := []slot.Slot{
		{Name: "fire", State: slot.SlotActive, Branch: "hotfix/y", HeadHash: "def5678", IsDirty: true},
	}
	result := RenderSlotList(slots, true)
	assert.Contains(t, result, "[dirty]")
	assert.Contains(t, result, "*dirty")
}

func TestRenderSlotList_NoIcons(t *testing.T) {
	slots := []slot.Slot{
		{Name: "slot-1"},
		{Name: "slot-2"},
	}
	result := RenderSlotList(slots, true)
	assert.NotContains(t, result, "  🌱")
	assert.Contains(t, result, "slot-1")
}

func TestRenderSlotList_MixedIcons(t *testing.T) {
	slots := []slot.Slot{
		{Name: "wood", Icon: "🌱"},
		{Name: "plain"},
	}
	result := RenderSlotList(slots, true)
	lines := strings.Split(result, "\n")
	assert.Contains(t, lines[0], "🌱")
}

func TestRenderSlotList_NameAlignment(t *testing.T) {
	slots := []slot.Slot{
		{Name: "a"},
		{Name: "long-name"},
	}
	result := RenderSlotList(slots, true)
	lines := strings.Split(result, "\n")
	idx0 := strings.Index(lines[0], "[empty]")
	idx1 := strings.Index(lines[1], "[empty]")
	assert.Equal(t, idx0, idx1, "state tags should be aligned")
}

func TestRenderSlotList_WithColor(t *testing.T) {
	slots := []slot.Slot{
		{Name: "wood", Icon: "🌱", State: slot.SlotActive, Branch: "main", HeadHash: "abc"},
	}
	result := RenderSlotList(slots, false)
	assert.Contains(t, result, "wood")
	assert.Contains(t, result, "main")
}
