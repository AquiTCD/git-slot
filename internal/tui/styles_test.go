package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestStateStyle_Active(t *testing.T) {
	s := StateStyle("active")
	assert.Equal(t, StyleStateActive, s)
}

func TestStateStyle_Dirty(t *testing.T) {
	s := StateStyle("dirty")
	assert.Equal(t, StyleStateDirty, s)
}

func TestStateStyle_Empty(t *testing.T) {
	s := StateStyle("empty")
	assert.Equal(t, StyleStateEmpty, s)
}

func TestStateStyle_Unknown(t *testing.T) {
	s := StateStyle("unknown")
	assert.Equal(t, StyleStateEmpty, s)
}

func TestRender_WithColor(t *testing.T) {
	result := render(StyleName, "hello", false)
	assert.Contains(t, result, "hello")
}

func TestRender_NoColor(t *testing.T) {
	result := render(StyleName, "hello", true)
	assert.Equal(t, "hello", result)
}

func TestRender_PreservesText(t *testing.T) {
	tests := []struct {
		name    string
		style   lipgloss.Style
		text    string
		noColor bool
	}{
		{"name style no-color", StyleName, "test-slot", true},
		{"branch style no-color", StyleBranch, "main", true},
		{"hash style no-color", StyleHash, "(abc1234)", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.text, render(tt.style, tt.text, tt.noColor))
		})
	}
}
