package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckExecAllowedInSlotShell(t *testing.T) {
	t.Run("outside shell always ok", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "")
		assert.NoError(t, checkExecAllowedInSlotShell(true, "other", "other"))
	})

	t.Run("inside shell explicit same slot", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "work")
		assert.NoError(t, checkExecAllowedInSlotShell(true, "work", "work"))
	})

	t.Run("inside shell explicit different slot", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "work")
		err := checkExecAllowedInSlotShell(true, "hotfix", "hotfix")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrShellNested))
	})

	t.Run("inside shell implicit cwd matches", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "work")
		assert.NoError(t, checkExecAllowedInSlotShell(false, "", "work"))
	})

	t.Run("inside shell implicit cwd mismatch", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "work")
		err := checkExecAllowedInSlotShell(false, "", "hotfix")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrShellNested))
	})
}
