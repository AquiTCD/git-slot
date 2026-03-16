package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AquiTCD/git-slot/internal/errutil"
)

var ErrNotInRepo = errutil.NewExitError("not inside a git repository", 3)

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

// MainRepoRoot returns the primary worktree root (for "eject").
// When inside a linked worktree this is the main repo path; otherwise same as RepoRoot().
func (d *ExecDetector) MainRepoRoot() (string, error) {
	toplevel, err := d.RepoRoot()
	if err != nil {
		return "", err
	}
	gitPath := filepath.Join(toplevel, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return "", ErrNotInRepo
	}
	if info.IsDir() {
		return toplevel, nil
	}
	// Worktree: .git is a file with "gitdir: <path>"
	raw, err := os.ReadFile(gitPath)
	if err != nil {
		return "", ErrNotInRepo
	}
	line := strings.TrimSpace(string(raw))
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return toplevel, nil
	}
	gitdir := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(toplevel, gitdir)
	}
	gitdir, err = filepath.Abs(gitdir)
	if err != nil {
		return "", ErrNotInRepo
	}
	// gitdir is e.g. .../main/.git/worktrees/slot-name
	worktreesSep := string(filepath.Separator) + "worktrees" + string(filepath.Separator)
	if idx := strings.Index(gitdir, worktreesSep); idx != -1 {
		mainGitDir := gitdir[:idx]
		return filepath.Dir(mainGitDir), nil
	}
	return toplevel, nil
}

func (d *ExecDetector) IsInsideRepo() bool {
	_, err := d.RepoRoot()
	return err == nil
}
