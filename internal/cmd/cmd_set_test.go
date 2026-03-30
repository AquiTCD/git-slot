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

// ENH-P1-T1: --pr flag is registered
func TestSetSubcommand_PRFlag_Registered(t *testing.T) {
	f := setCmd.Flags().Lookup("pr")
	require.NotNil(t, f, "--pr flag should be registered on set command")
}

// ENH-P1-T2: --pr and --create cannot be used together
func TestSetSubcommand_PRAndCreate_Error(t *testing.T) {
	_, err := executeCommand("set", "work", "--pr", "1", "-c", "new-branch")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--pr と --create は同時に使えません")
}
