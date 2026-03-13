package tui

import (
	"testing"

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
