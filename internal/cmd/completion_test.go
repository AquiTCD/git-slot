package cmd

import (
	"strings"
	"testing"
)

func TestCompletion_Bash(t *testing.T) {
	out, err := executeCommand("completion", "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bash") {
		t.Errorf("expected bash completion output, got: %s", out)
	}
}

func TestCompletion_Zsh(t *testing.T) {
	out, err := executeCommand("completion", "zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected zsh completion output, got empty string")
	}
}

func TestCompletion_Fish(t *testing.T) {
	out, err := executeCommand("completion", "fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected fish completion output, got empty string")
	}
}

func TestCompletion_PowerShell(t *testing.T) {
	out, err := executeCommand("completion", "powershell")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected powershell completion output, got empty string")
	}
}

func TestCompletion_InvalidShell(t *testing.T) {
	_, err := executeCommand("completion", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid shell argument")
	}
}

func TestCompletion_NoArgs(t *testing.T) {
	_, err := executeCommand("completion")
	if err == nil {
		t.Fatal("expected error when no shell argument provided")
	}
}

func TestWrapper_Zsh(t *testing.T) {
	out, err := executeCommand("wrapper", "zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "gs()") {
		t.Errorf("expected gs function definition, got: %s", out)
	}
	if !strings.Contains(out, "cd") {
		t.Errorf("expected cd in wrapper, got: %s", out)
	}
}

func TestWrapper_Bash(t *testing.T) {
	out, err := executeCommand("wrapper", "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "gs()") {
		t.Errorf("expected gs function definition, got: %s", out)
	}
}

func TestWrapper_Fish(t *testing.T) {
	out, err := executeCommand("wrapper", "fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "function gs") {
		t.Errorf("expected fish gs function, got: %s", out)
	}
}

func TestWrapper_InvalidShell(t *testing.T) {
	_, err := executeCommand("wrapper", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid shell argument")
	}
}

func TestWrapper_NoArgs(t *testing.T) {
	_, err := executeCommand("wrapper")
	if err == nil {
		t.Fatal("expected error when no shell argument provided")
	}
}
