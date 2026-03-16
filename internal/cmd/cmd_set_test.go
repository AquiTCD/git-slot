package cmd

import (
	"testing"
)

func TestSetSubcommand_NoArgs(t *testing.T) {
	_, err := executeCommand("set")
	if err == nil {
		t.Fatal("expected error when no slot name provided")
	}
}

func TestSetSubcommand_TooManyArgs(t *testing.T) {
	_, err := executeCommand("set", "slot", "branch", "extra")
	if err == nil {
		t.Fatal("expected error for too many arguments")
	}
}

func TestSetSubcommand_ForceShortFlag(t *testing.T) {
	f := setCmd.Flags().Lookup("force")
	if f == nil {
		t.Fatal("expected 'force' flag on set subcommand")
	}
	if f.Shorthand != "f" {
		t.Errorf("expected shorthand 'f' for 'force' flag, got '%s'", f.Shorthand)
	}
}

func TestSetSubcommand_CreateShortFlag(t *testing.T) {
	f := setCmd.Flags().Lookup("create")
	if f == nil {
		t.Fatal("expected 'create' flag on set subcommand")
	}
	if f.Shorthand != "c" {
		t.Errorf("expected shorthand 'c' for 'create' flag, got '%s'", f.Shorthand)
	}
}

func TestSetSubcommand_BranchShortFlag(t *testing.T) {
	f := setCmd.Flags().Lookup("branch")
	if f == nil {
		t.Fatal("expected 'branch' flag on set subcommand")
	}
	if f.Shorthand != "b" {
		t.Errorf("expected shorthand 'b' for 'branch' flag, got '%s'", f.Shorthand)
	}
}
