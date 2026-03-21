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

func TestExecWorktree_Switch(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "branch-a")
	run(t, dir, "git", "branch", "branch-b")

	wtPath := filepath.Join(t.TempDir(), "wt-switch")
	w := git.NewExecWorktree(dir)
	require.NoError(t, w.Add(wtPath, "branch-a"))

	err := w.Switch(wtPath, "branch-b", false)
	require.NoError(t, err)

	entries, err := w.List()
	require.NoError(t, err)

	resolved, _ := filepath.EvalSymlinks(wtPath)
	var found bool
	for _, e := range entries {
		if e.Path == resolved {
			assert.Equal(t, "branch-b", e.Branch)
			found = true
		}
	}
	assert.True(t, found, "worktree should still exist after switch")
}

func TestExecWorktree_Switch_PreservesUntrackedFiles(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "sw-a")
	run(t, dir, "git", "branch", "sw-b")

	wtPath := filepath.Join(t.TempDir(), "wt-preserve")
	w := git.NewExecWorktree(dir)
	require.NoError(t, w.Add(wtPath, "sw-a"))

	untrackedFile := filepath.Join(wtPath, "node_modules", "pkg.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(untrackedFile), 0o755))
	require.NoError(t, os.WriteFile(untrackedFile, []byte("{}"), 0o644))

	err := w.Switch(wtPath, "sw-b", false)
	require.NoError(t, err)

	_, err = os.Stat(untrackedFile)
	assert.NoError(t, err, "untracked files should survive switch")
}

func TestExecWorktree_Switch_BranchInUse(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "shared")
	run(t, dir, "git", "branch", "other")

	wt1 := filepath.Join(t.TempDir(), "wt1")
	wt2 := filepath.Join(t.TempDir(), "wt2")
	w := git.NewExecWorktree(dir)
	require.NoError(t, w.Add(wt1, "shared"))
	require.NoError(t, w.Add(wt2, "other"))

	err := w.Switch(wt2, "shared", false)
	assert.Error(t, err, "switching to a branch checked out elsewhere should fail")
}

func TestExecWorktree_SwitchCreate(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "base")

	wtPath := filepath.Join(t.TempDir(), "wt-create")
	w := git.NewExecWorktree(dir)
	require.NoError(t, w.Add(wtPath, "base"))

	err := w.SwitchCreate(wtPath, "brand-new")
	require.NoError(t, err)

	entries, err := w.List()
	require.NoError(t, err)

	resolved, _ := filepath.EvalSymlinks(wtPath)
	var found bool
	for _, e := range entries {
		if e.Path == resolved {
			assert.Equal(t, "brand-new", e.Branch)
			found = true
		}
	}
	assert.True(t, found, "worktree should show new branch after switch -c")

	exists, err := w.BranchExists("brand-new")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExecWorktree_SwitchCreate_AlreadyExists(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "start")
	run(t, dir, "git", "branch", "already-here")

	wtPath := filepath.Join(t.TempDir(), "wt-dup")
	w := git.NewExecWorktree(dir)
	require.NoError(t, w.Add(wtPath, "start"))

	err := w.SwitchCreate(wtPath, "already-here")
	assert.Error(t, err, "switch -c to existing branch should fail")
}

func TestExecWorktree_Switch_DiscardChanges(t *testing.T) {
	dir := setupTestRepo(t)
	run(t, dir, "git", "branch", "dc-a")
	run(t, dir, "git", "branch", "dc-b")

	wtPath := filepath.Join(t.TempDir(), "wt-discard")
	w := git.NewExecWorktree(dir)
	require.NoError(t, w.Add(wtPath, "dc-a"))

	require.NoError(t, os.WriteFile(filepath.Join(wtPath, "tracked.txt"), []byte("data"), 0o644))
	run(t, wtPath, "git", "add", "tracked.txt")
	run(t, wtPath, "git", "commit", "-m", "add tracked")
	require.NoError(t, os.WriteFile(filepath.Join(wtPath, "tracked.txt"), []byte("modified"), 0o644))

	err := w.Switch(wtPath, "dc-b", false)
	assert.Error(t, err, "switch with dirty tracked files should fail without discard")

	err = w.Switch(wtPath, "dc-b", true)
	require.NoError(t, err, "switch with --discard-changes should succeed")
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

// F2-G1: DirtyCount returns the number of dirty files
func TestExecWorktree_DirtyCount_WithDirtyFiles(t *testing.T) {
	dir := setupTestRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("data"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("data"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file3.txt"), []byte("data"), 0o644))

	w := git.NewExecWorktree(dir)
	count, err := w.DirtyCount(dir)

	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestExecWorktree_DirtyCount_Clean(t *testing.T) {
	dir := setupTestRepo(t)
	w := git.NewExecWorktree(dir)

	count, err := w.DirtyCount(dir)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// F2-G2: AheadCount returns the number of commits ahead of upstream
func TestExecWorktree_AheadCount_WithAheadCommits(t *testing.T) {
	dir := setupTestRepo(t)

	// Set up a remote-like branch as upstream by creating a "remote" branch
	// then add commits on top of it
	run(t, dir, "git", "branch", "upstream-ref")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "commit1.txt"), []byte("c1"), 0o644))
	run(t, dir, "git", "add", "commit1.txt")
	run(t, dir, "git", "commit", "-m", "commit 1")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "commit2.txt"), []byte("c2"), 0o644))
	run(t, dir, "git", "add", "commit2.txt")
	run(t, dir, "git", "commit", "-m", "commit 2")

	// Configure upstream tracking
	currentBranch := run(t, dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	run(t, dir, "git", "config", "branch."+currentBranch+".remote", ".")
	run(t, dir, "git", "config", "branch."+currentBranch+".merge", "refs/heads/upstream-ref")

	w := git.NewExecWorktree(dir)
	count, err := w.AheadCount(dir)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// F2-G3: AheadCount returns error when upstream is not configured
func TestExecWorktree_AheadCount_NoUpstream(t *testing.T) {
	dir := setupTestRepo(t)
	w := git.NewExecWorktree(dir)

	_, err := w.AheadCount(dir)

	assert.Error(t, err, "should return error when upstream is not configured")
}
