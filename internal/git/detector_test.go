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
