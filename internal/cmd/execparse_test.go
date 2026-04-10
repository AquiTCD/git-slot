package cmd

import (
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseExecArgv(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}, {Name: "hotfix"}},
	}

	t.Run("no slot", func(t *testing.T) {
		explicit, slot, cmd, err := parseExecArgv([]string{"git-slot", "exec", "--", "echo", "hi"}, cfg)
		require.NoError(t, err)
		assert.False(t, explicit)
		assert.Equal(t, "", slot)
		assert.Equal(t, []string{"echo", "hi"}, cmd)
	})

	t.Run("with slot", func(t *testing.T) {
		explicit, slot, cmd, err := parseExecArgv([]string{"git-slot", "exec", "work", "--", "npm", "test"}, cfg)
		require.NoError(t, err)
		assert.True(t, explicit)
		assert.Equal(t, "work", slot)
		assert.Equal(t, []string{"npm", "test"}, cmd)
	})

	t.Run("missing dash dash", func(t *testing.T) {
		_, _, _, err := parseExecArgv([]string{"git-slot", "exec", "echo"}, cfg)
		require.Error(t, err)
	})

	t.Run("unknown slot", func(t *testing.T) {
		_, _, _, err := parseExecArgv([]string{"git-slot", "exec", "nope", "--", "true"}, cfg)
		require.Error(t, err)
	})

	t.Run("too many before separator", func(t *testing.T) {
		_, _, _, err := parseExecArgv([]string{"git-slot", "exec", "a", "b", "--", "true"}, cfg)
		require.Error(t, err)
	})
}

func TestExecWantsHelp(t *testing.T) {
	assert.True(t, execWantsHelp([]string{"git-slot", "exec", "--help"}))
	assert.True(t, execWantsHelp([]string{"git-slot", "exec", "-h"}))
	assert.False(t, execWantsHelp([]string{"git-slot", "exec", "work", "--", "npm", "--help"}))
	assert.False(t, execWantsHelp([]string{"git-slot", "exec", "--", "sh", "-c", "exit 0"}))
}
