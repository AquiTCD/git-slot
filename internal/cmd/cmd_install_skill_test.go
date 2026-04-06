package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallSkill(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
		wantErr bool
	}{
		{
			name:    "stdout flag",
			args:    []string{"install-skill", "--stdout"},
			wantOut: "Git-Slot: Parallel Worktree Management",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd

			stdout := new(bytes.Buffer)
			cmd.SetOut(stdout)
			cmd.SetErr(new(bytes.Buffer))

			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantOut != "" {
				got := stdout.String()
				if !strings.Contains(got, tt.wantOut) {
					t.Errorf("expected stdout to contain %q, got: %v", tt.wantOut, got)
				}
			}
		})
	}
}

func TestInstallSkill_Files(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-slot-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	t.Run("custom dir", func(t *testing.T) {
		cmd := rootCmd
		_ = installSkillCmd.Flags().Set("stdout", "false")
		_ = installSkillCmd.Flags().Set("append", "")
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs([]string{"install-skill", "my-agents/skill"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFile := filepath.Join("my-agents", "skill", "git-slot-workflow", "SKILL.md")
		b, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("failed to read created file %q: %v", expectedFile, err)
		}

		if !strings.Contains(string(b), "Git-Slot: Parallel Worktree Management") {
			t.Errorf("file did not contain expected content, got: %s", string(b))
		}
	})

	t.Run("append to file", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, ".cursorrules")
		_ = os.WriteFile(targetFile, []byte("existing rule\n"), 0644)

		cmd := rootCmd
		_ = installSkillCmd.Flags().Set("stdout", "false")
		_ = installSkillCmd.Flags().Set("append", targetFile)
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs([]string{"install-skill", "--append", targetFile})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		content := string(b)
		if !strings.Contains(content, "existing rule") {
			t.Errorf(".cursorrules is missing existing content")
		}
		if !strings.Contains(content, "Git-Slot: Parallel Worktree Management") {
			t.Errorf(".cursorrules is missing new appended content")
		}
	})
}
