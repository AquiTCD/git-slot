package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// IgnoredFile represents a file that is ignored by git
type IgnoredFile struct {
	Path string
}

// ListIgnoredFiles returns a list of files that are ignored by git
func (w *ExecWorktree) ListIgnoredFiles() ([]IgnoredFile, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--ignored", "--exclude-standard", "--directory")
	cmd.Dir = w.dir

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	var ignored []IgnoredFile
	for _, line := range lines {
		if line == "" {
			continue
		}
		ignored = append(ignored, IgnoredFile{Path: line})
	}
	return ignored, nil
}

// IsIgnored checks if the given path is ignored by git
func (w *ExecWorktree) IsIgnored(path string) (bool, error) {
	cmd := exec.Command("git", "check-ignore", "-q", path)
	cmd.Dir = w.dir

	err := cmd.Run()
	if err == nil {
		return true, nil // exit code 0 means ignored
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 1 {
			return false, nil // exit code 1 means NOT ignored
		}
	}

	return false, fmt.Errorf("git check-ignore: %w", err)
}
