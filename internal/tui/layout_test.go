package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterFieldWidth_DefaultTermWidth(t *testing.T) {
	assert.Equal(t, 82, filterFieldWidth(0))
}

func TestFilterFieldWidth_TracksTerminal(t *testing.T) {
	assert.Equal(t, max(8, 120-6), filterFieldWidth(120))
	assert.Equal(t, 8, filterFieldWidth(10))
}
