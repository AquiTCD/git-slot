package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})

	resetSubcommandFlags(rootCmd)

	err := rootCmd.Execute()
	return buf.String(), err
}

func resetSubcommandFlags(cmd *cobra.Command) {
	for _, sub := range cmd.Commands() {
		sub.Flags().VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
		resetSubcommandFlags(sub)
	}
}

func TestVersionFlag(t *testing.T) {
	version = "0.0.0-test"
	commit = "abc1234"
	date = "2026-01-01"

	out, err := executeCommand("--version")
	require.NoError(t, err)
	assert.Contains(t, out, "git-slot version 0.0.0-test")
	assert.Contains(t, out, "abc1234")
}

func TestVersionShortFlag(t *testing.T) {
	version = "0.0.0-test"
	commit = "abc1234"
	date = "2026-01-01"

	out, err := executeCommand("-v")
	require.NoError(t, err)
	assert.Contains(t, out, "git-slot version 0.0.0-test")
}

func TestHelpWithNoArgs(t *testing.T) {
	out, err := executeCommand()
	require.NoError(t, err)
	assert.Contains(t, out, "git-slot")
	assert.Contains(t, out, "set")
	assert.Contains(t, out, "list")
}

func TestUnknownSubcommand(t *testing.T) {
	_, err := executeCommand("nonexistent-cmd")
	require.Error(t, err)
}

func TestGlobalShorthandOnInit(t *testing.T) {
	f := initCmd.Flags().Lookup("global")
	require.NotNil(t, f)
	assert.Equal(t, "g", f.Shorthand)
}

func TestGlobalShorthandOnHook(t *testing.T) {
	f := hookCmd.Flags().Lookup("global")
	require.NotNil(t, f)
	assert.Equal(t, "g", f.Shorthand)
}
