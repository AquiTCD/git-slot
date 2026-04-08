package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeSlotSelectLayout_DefaultTermWidth(t *testing.T) {
	lay := computeSlotSelectLayout(0)
	assert.Equal(t, 84, lay.Inner)
	assert.Equal(t, 51, lay.Left)
	assert.Equal(t, 32, lay.Right)
	assert.Equal(t, 50, lay.LeftContent)
	assert.Equal(t, 31, lay.RightContent)
	assert.Equal(t, lay.Inner, lay.Left+1+lay.Right, "inner must equal left + │ + right")
}

func TestComputeSlotSelectLayout_InnerEqualsSumOfColumns(t *testing.T) {
	for _, tw := range []int{40, 60, 72, 88, 100, 120, 200, 300} {
		t.Run(fmt.Sprintf("w%d", tw), func(t *testing.T) {
			lay := computeSlotSelectLayout(tw)
			assert.Equal(t, lay.Inner, lay.Left+1+lay.Right, "termWidth=%d", tw)
			assert.Positive(t, lay.LeftContent)
			assert.GreaterOrEqual(t, lay.RightContent, 8)
		})
	}
}

func TestComputeSlotSelectLayout_LeftNotSmallerThanRight(t *testing.T) {
	lay := computeSlotSelectLayout(120)
	assert.GreaterOrEqual(t, lay.Left, lay.Right, "slot column should be the wider pane")
}

func TestFilterFieldWidth_TracksLeftContent(t *testing.T) {
	lay := computeSlotSelectLayout(120)
	assert.Equal(t, max(8, lay.LeftContent), filterFieldWidth(120))
}

func TestJoinSplitPaneRows_EachRowMatchesInnerWidth(t *testing.T) {
	lay := computeSlotSelectLayout(120)
	m := newInteractiveTestModel(testSlots(), true)
	m.width = 120
	leftBlock := m.buildLeftPane(lay.LeftContent, lay.Left)
	rightBlock := m.buildRightPane(lay.RightContent, lay.Right)
	rows := joinSplitPaneRows(leftBlock, rightBlock, "│", lay)
	sepW := ansi.StringWidth("│")
	require.NotEmpty(t, rows)
	for i, row := range rows {
		assert.Equal(t, lay.Inner, ansi.StringWidth(row), "row %d", i)
		assert.Equal(t, sepW, 1)
		parts := strings.SplitN(row, "│", 2)
		require.Len(t, parts, 2, "row %d", i)
		assert.Equal(t, lay.Left, ansi.StringWidth(parts[0]), "left col row %d", i)
		assert.Equal(t, lay.Right, ansi.StringWidth(parts[1]), "right col row %d", i)
	}
}

func TestFormatGridColumnLine_FixedWidth(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		content int
		col     int
		wantW   int
	}{
		{"empty text pads", "", 12, 14, 14},
		{"short pads right", "hi", 10, 12, 12},
		{"long truncates", strings.Repeat("x", 200), 8, 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatGridColumnLine(tt.line, tt.content, tt.col)
			assert.Equal(t, tt.wantW, ansi.StringWidth(got))
		})
	}
}
