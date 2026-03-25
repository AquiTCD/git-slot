package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSubcommand_NoArgs(t *testing.T) {
	_, err := executeCommand("set")
	require.Error(t, err)
}

func TestSetSubcommand_TooManyArgs(t *testing.T) {
	_, err := executeCommand("set", "slot", "branch", "extra")
	require.Error(t, err)
}

func TestSetSubcommand_ForceShortFlag(t *testing.T) {
	f := setCmd.Flags().Lookup("force")
	require.NotNil(t, f)
	assert.Equal(t, "f", f.Shorthand)
}

func TestSetSubcommand_CreateShortFlag(t *testing.T) {
	f := setCmd.Flags().Lookup("create")
	require.NotNil(t, f)
	assert.Equal(t, "c", f.Shorthand)
}

func TestSetSubcommand_BranchShortFlag(t *testing.T) {
	f := setCmd.Flags().Lookup("branch")
	require.NotNil(t, f)
	assert.Equal(t, "b", f.Shorthand)
}
