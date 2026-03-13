package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type WorktreeInfo struct {
	Path     string
	Branch   string
	HeadHash string
	IsBare   bool
}

type Worktree interface {
	List() ([]WorktreeInfo, error)
	Add(path, branch string) error
	AddNewBranch(path, newBranch string) error
	Remove(path string, force bool) error
	Move(oldPath, newPath string) error
	BranchExists(branch string) (bool, error)
	IsDirty(path string) (bool, error)
	CommitSubject(path string) (string, error)
	StatusShort(path string) ([]string, error)
	ListBranches() ([]string, error)
}

type ExecWorktree struct {
	dir string
}

var _ Worktree = (*ExecWorktree)(nil)

func NewExecWorktree(dir string) *ExecWorktree {
	return &ExecWorktree{dir: dir}
}

func (w *ExecWorktree) List() ([]WorktreeInfo, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = w.dir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	var entries []WorktreeInfo
	var current *WorktreeInfo

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "worktree "):
			if current != nil {
				entries = append(entries, *current)
			}
			current = &WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}

		case strings.HasPrefix(line, "HEAD "):
			if current != nil {
				hash := strings.TrimPrefix(line, "HEAD ")
				if len(hash) > 7 {
					hash = hash[:7]
				}
				current.HeadHash = hash
			}

		case strings.HasPrefix(line, "branch "):
			if current != nil {
				current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
			}

		case line == "bare":
			if current != nil {
				current.IsBare = true
			}
		}
	}

	if current != nil {
		entries = append(entries, *current)
	}

	return entries, nil
}

func (w *ExecWorktree) Add(path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", path, branch)
	cmd.Dir = w.dir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func (w *ExecWorktree) AddNewBranch(path, newBranch string) error {
	cmd := exec.Command("git", "worktree", "add", path, "-b", newBranch)
	cmd.Dir = w.dir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add -b: %s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func (w *ExecWorktree) Remove(path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = w.dir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove: %s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func (w *ExecWorktree) BranchExists(branch string) (bool, error) {
	localCmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch)
	localCmd.Dir = w.dir
	if err := localCmd.Run(); err == nil {
		return true, nil
	}

	remoteCmd := exec.Command("git", "rev-parse", "--verify", "refs/remotes/origin/"+branch)
	remoteCmd.Dir = w.dir
	if err := remoteCmd.Run(); err == nil {
		return true, nil
	}

	return false, nil
}

func (w *ExecWorktree) IsDirty(path string) (bool, error) {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")

	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}

	return strings.TrimSpace(string(out)) != "", nil
}

func (w *ExecWorktree) Move(oldPath, newPath string) error {
	cmd := exec.Command("git", "worktree", "move", oldPath, newPath)
	cmd.Dir = w.dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree move: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func (w *ExecWorktree) CommitSubject(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "log", "-1", "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (w *ExecWorktree) ListBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--list", "--format=%(refname:short)")
	cmd.Dir = w.dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch --list: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

func (w *ExecWorktree) StatusShort(path string) ([]string, error) {
	cmd := exec.Command("git", "-C", path, "status", "--short")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status --short: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}
