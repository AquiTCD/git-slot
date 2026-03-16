package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func TestVersionShortFlag(t *testing.T) {
	version = "0.0.0-test"
	commit = "abc1234"
	date = "2026-01-01"

	out, err := executeCommand("-v")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "git-slot version 0.0.0-test") {
		t.Errorf("expected version output, got: %s", out)
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
	if !strings.Contains(out, "set") {
		t.Errorf("expected 'set' subcommand in help, got: %s", out)
	}
	if !strings.Contains(out, "list") {
		t.Errorf("expected 'list' subcommand in help, got: %s", out)
	}
}

func TestUnknownSubcommand(t *testing.T) {
	_, err := executeCommand("nonexistent-cmd")
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}

func TestGlobalShorthandOnInit(t *testing.T) {
	f := initCmd.Flags().Lookup("global")
	if f == nil {
		t.Fatal("expected 'global' flag on init subcommand")
	}
	if f.Shorthand != "g" {
		t.Errorf("expected shorthand 'g' for 'global' flag, got '%s'", f.Shorthand)
	}
}

func TestGlobalShorthandOnHook(t *testing.T) {
	f := hookCmd.Flags().Lookup("global")
	if f == nil {
		t.Fatal("expected 'global' flag on hook subcommand")
	}
	if f.Shorthand != "g" {
		t.Errorf("expected shorthand 'g' for 'global' flag, got '%s'", f.Shorthand)
	}
}
