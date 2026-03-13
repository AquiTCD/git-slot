package tui

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTTY_Buffer(t *testing.T) {
	buf := new(bytes.Buffer)
	assert.False(t, IsTTY(buf))
}

func TestIsTTY_DevNull(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Skip("cannot open /dev/null")
	}
	defer f.Close()
	assert.False(t, IsTTY(f))
}

func TestIsNoColor_Unset(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	assert.True(t, IsNoColor())
}

func TestIsNoColor_Set(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	assert.True(t, IsNoColor())
}
