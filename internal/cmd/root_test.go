package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	// Reset all flags to defaults before each invocation
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})

	err := rootCmd.Execute()
	return buf.String(), err
}

func TestVersionFlag(t *testing.T) {
	version = "0.0.0-test"
	commit = "abc1234"
	date = "2026-01-01"

	out, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "git-slot version 0.0.0-test") {
		t.Errorf("expected version output, got: %s", out)
	}
	if !strings.Contains(out, "abc1234") {
		t.Errorf("expected commit hash in output, got: %s", out)
	}
}

func TestHelpWithNoArgs(t *testing.T) {
	out, err := executeCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "git-slot") {
		t.Errorf("expected help output, got: %s", out)
	}
}

func TestTooManyArgs(t *testing.T) {
	_, err := executeCommand("a", "b", "c")
	if err == nil {
		t.Fatal("expected error for too many arguments")
	}
}
