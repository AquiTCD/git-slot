package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletion_Bash(t *testing.T) {
	out, err := executeCommand("completion", "bash")
	require.NoError(t, err)
	assert.Contains(t, out, "bash")
}

func TestCompletion_Zsh(t *testing.T) {
	out, err := executeCommand("completion", "zsh")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletion_Fish(t *testing.T) {
	out, err := executeCommand("completion", "fish")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletion_PowerShell(t *testing.T) {
	out, err := executeCommand("completion", "powershell")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletion_InvalidShell(t *testing.T) {
	_, err := executeCommand("completion", "invalid")
	require.Error(t, err)
}

func TestCompletion_NoArgs(t *testing.T) {
	_, err := executeCommand("completion")
	require.Error(t, err)
}

func TestWrapper_Zsh(t *testing.T) {
	out, err := executeCommand("wrapper", "zsh")
	require.NoError(t, err)
	assert.Contains(t, out, "gsl()")
	assert.Contains(t, out, "cd")
}

func TestWrapper_Bash(t *testing.T) {
	out, err := executeCommand("wrapper", "bash")
	require.NoError(t, err)
	assert.Contains(t, out, "gsl()")
}

func TestWrapper_Fish(t *testing.T) {
	out, err := executeCommand("wrapper", "fish")
	require.NoError(t, err)
	assert.Contains(t, out, "function gsl")
}

func TestWrapper_InvalidShell(t *testing.T) {
	_, err := executeCommand("wrapper", "invalid")
	require.Error(t, err)
}

func TestWrapper_NoArgs(t *testing.T) {
	_, err := executeCommand("wrapper")
	require.Error(t, err)
}

func TestWrapper_NoDoubleExecution(t *testing.T) {
	shells := []string{"zsh", "bash", "fish"}
	for _, shell := range shells {
		out, err := executeCommand("wrapper", shell)
		require.NoError(t, err, "shell: %s", shell)

		count := strings.Count(out, "command git-slot")
		assert.Equal(t, 3, count, "expected 'command git-slot' to appear exactly 3 times in %s wrapper", shell)
	}
}

func TestWrapper_ShellSubcommandBypass(t *testing.T) {
	shells := []string{"zsh", "bash", "fish"}
	for _, shell := range shells {
		out, err := executeCommand("wrapper", shell)
		require.NoError(t, err, "shell: %s", shell)
		assert.Contains(t, out, "shell", "expected shell subcommand guard in %s wrapper", shell)
	}
}

func TestWrapper_SetsGslFromWrapperForCapturedInvocations(t *testing.T) {
	shells := []string{"zsh", "bash", "fish"}
	for _, shell := range shells {
		out, err := executeCommand("wrapper", shell)
		require.NoError(t, err, "shell: %s", shell)
		assert.Contains(t, out, "GSL_FROM_WRAPPER=1", "shell: %s", shell)
	}
}
