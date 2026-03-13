package git

import (
	"errors"
	"os/exec"
	"strings"
)

var ErrNotInRepo = errors.New("not inside a git repository")

type Detector interface {
	RepoRoot() (string, error)
	IsInsideRepo() bool
}

type ExecDetector struct {
	dir string
}

func NewExecDetector(dir string) *ExecDetector {
	return &ExecDetector{dir: dir}
}

func (d *ExecDetector) RepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if d.dir != "" {
		cmd.Dir = d.dir
	}

	out, err := cmd.Output()
	if err != nil {
		return "", ErrNotInRepo
	}

	return strings.TrimSpace(string(out)), nil
}

func (d *ExecDetector) IsInsideRepo() bool {
	_, err := d.RepoRoot()
	return err == nil
}
