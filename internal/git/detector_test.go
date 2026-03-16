package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", out)
}

func TestRepoRoot_InsideRepo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	d := git.NewExecDetector(dir)
	root, err := d.RepoRoot()

	require.NoError(t, err)
	resolved, _ := filepath.EvalSymlinks(dir)
	assert.Equal(t, resolved, root)
}

func TestRepoRoot_OutsideRepo(t *testing.T) {
	dir := t.TempDir()

	d := git.NewExecDetector(dir)
	_, err := d.RepoRoot()

	assert.ErrorIs(t, err, git.ErrNotInRepo)
}

func TestIsInsideRepo_True(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	d := git.NewExecDetector(dir)
	assert.True(t, d.IsInsideRepo())
}

func TestIsInsideRepo_False(t *testing.T) {
	dir := t.TempDir()

	d := git.NewExecDetector(dir)
	assert.False(t, d.IsInsideRepo())
}

func TestRepoRoot_FromSubdirectory(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	sub := filepath.Join(dir, "a", "b")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	d := git.NewExecDetector(sub)
	root, err := d.RepoRoot()

	require.NoError(t, err)
	resolved, _ := filepath.EvalSymlinks(dir)
	assert.Equal(t, resolved, root)
}

func TestMainRepoRoot_MainWorktree(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	d := git.NewExecDetector(dir)
	mainRoot, err := d.MainRepoRoot()

	require.NoError(t, err)
	repoRoot, _ := d.RepoRoot()
	assert.Equal(t, repoRoot, mainRoot)
	resolved, _ := filepath.EvalSymlinks(dir)
	assert.Equal(t, resolved, mainRoot)
}

func TestMainRepoRoot_LinkedWorktree(t *testing.T) {
	mainDir := t.TempDir()
	initGitRepo(t, mainDir)
	// Need at least one commit for worktree add
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "init")
	cmd.Dir = mainDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git commit: %s", out)

	wtDir := t.TempDir()
	cmd = exec.Command("git", "worktree", "add", wtDir, "-b", "wt-test")
	cmd.Dir = mainDir
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "git worktree add: %s", out)
	t.Cleanup(func() {
		rm := exec.Command("git", "worktree", "remove", wtDir)
		rm.Dir = mainDir
		_ = rm.Run()
	})

	d := git.NewExecDetector(wtDir)
	mainRoot, err := d.MainRepoRoot()

	require.NoError(t, err)
	mainResolved, _ := filepath.EvalSymlinks(mainDir)
	assert.Equal(t, mainResolved, mainRoot)

	wtRoot, err := d.RepoRoot()
	require.NoError(t, err)
	wtResolved, _ := filepath.EvalSymlinks(wtDir)
	assert.Equal(t, wtResolved, wtRoot)
	assert.NotEqual(t, wtRoot, mainRoot)
}
