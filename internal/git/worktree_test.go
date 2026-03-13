package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "initial")
	return dir
}

func run(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "command %s %v failed: %s", name, args, string(out))
	return strings.TrimSpace(string(out))
}

func TestExecWorktree_List_MainOnly(t *testing.T) {
	dir := setupTestRepo(t)
	w := git.NewExecWorktree(dir)

	entries, err := w.List()
	require.NoError(t, err)
	require.Len(t, entries, 1)

	resolved, _ := filepath.EvalSymlinks(dir)
	assert.Equal(t, resolved, entries[0].Path)
	assert.NotEmpty(t, entries[0].Branch)
	assert.NotEmpty(t, entries[0].HeadHash)
	assert.Len(t, entries[0].HeadHash, 7)
}

func TestExecWorktree_List_WithWorktree(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "feature-a")

	wtPath := filepath.Join(t.TempDir(), "wt-feature-a")
	w := git.NewExecWorktree(dir)

	require.NoError(t, w.Add(wtPath, "feature-a"))

	entries, err := w.List()
	require.NoError(t, err)
	require.Len(t, entries, 2)

	branches := []string{entries[0].Branch, entries[1].Branch}
	assert.Contains(t, branches, "feature-a")
}

func TestExecWorktree_Add(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "feature-b")

	wtPath := filepath.Join(t.TempDir(), "wt-feature-b")
	w := git.NewExecWorktree(dir)

	err := w.Add(wtPath, "feature-b")
	require.NoError(t, err)

	info, err := os.Stat(wtPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestExecWorktree_AddNewBranch(t *testing.T) {
	dir := setupTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "wt-new-branch")
	w := git.NewExecWorktree(dir)

	err := w.AddNewBranch(wtPath, "new-feature")
	require.NoError(t, err)

	info, err := os.Stat(wtPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	exists, err := w.BranchExists("new-feature")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExecWorktree_AddNewBranch_AlreadyExists(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "existing")

	wtPath := filepath.Join(t.TempDir(), "wt-existing")
	w := git.NewExecWorktree(dir)

	err := w.AddNewBranch(wtPath, "existing")
	assert.Error(t, err)
}

func TestExecWorktree_Remove(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "to-remove")

	wtPath := filepath.Join(t.TempDir(), "wt-remove")
	w := git.NewExecWorktree(dir)

	require.NoError(t, w.Add(wtPath, "to-remove"))

	err := w.Remove(wtPath, false)
	require.NoError(t, err)

	_, err = os.Stat(wtPath)
	assert.True(t, os.IsNotExist(err))
}

func TestExecWorktree_Remove_Force(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "dirty-branch")

	wtPath := filepath.Join(t.TempDir(), "wt-dirty")
	w := git.NewExecWorktree(dir)

	require.NoError(t, w.Add(wtPath, "dirty-branch"))
	require.NoError(t, os.WriteFile(filepath.Join(wtPath, "dirty.txt"), []byte("dirty"), 0o644))

	err := w.Remove(wtPath, true)
	require.NoError(t, err)
}

func TestExecWorktree_BranchExists_True(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "exists-branch")

	w := git.NewExecWorktree(dir)
	exists, err := w.BranchExists("exists-branch")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExecWorktree_BranchExists_False(t *testing.T) {
	dir := setupTestRepo(t)
	w := git.NewExecWorktree(dir)

	exists, err := w.BranchExists("no-such-branch")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestExecWorktree_IsDirty_Clean(t *testing.T) {
	dir := setupTestRepo(t)
	w := git.NewExecWorktree(dir)

	dirty, err := w.IsDirty(dir)

	require.NoError(t, err)
	assert.False(t, dirty)
}

func TestExecWorktree_IsDirty_Dirty(t *testing.T) {
	dir := setupTestRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hello"), 0o644))

	w := git.NewExecWorktree(dir)
	dirty, err := w.IsDirty(dir)

	require.NoError(t, err)
	assert.True(t, dirty)
}

func TestExecWorktree_Move(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "move-me")

	oldPath := filepath.Join(t.TempDir(), "wt-old")
	w := git.NewExecWorktree(dir)
	require.NoError(t, w.Add(oldPath, "move-me"))

	newPath := filepath.Join(t.TempDir(), "wt-new")
	err := w.Move(oldPath, newPath)
	require.NoError(t, err)

	_, err = os.Stat(oldPath)
	assert.True(t, os.IsNotExist(err))

	info, err := os.Stat(newPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestExecWorktree_CommitSubject(t *testing.T) {
	dir := setupTestRepo(t)
	w := git.NewExecWorktree(dir)

	subject, err := w.CommitSubject(dir)
	require.NoError(t, err)
	assert.Equal(t, "initial", subject)
}

func TestExecWorktree_StatusShort_Clean(t *testing.T) {
	dir := setupTestRepo(t)
	w := git.NewExecWorktree(dir)

	lines, err := w.StatusShort(dir)
	require.NoError(t, err)
	assert.Nil(t, lines)
}

func TestExecWorktree_StatusShort_Dirty(t *testing.T) {
	dir := setupTestRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("data"), 0o644))

	w := git.NewExecWorktree(dir)
	lines, err := w.StatusShort(dir)

	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], "untracked.txt")
}
