package git_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRepoWithIgnored(t *testing.T) string {
	t.Helper()
	dir := setupTestRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored.txt\nsecret/\n"), 0644))
	run(t, dir, "git", "add", ".gitignore")
	run(t, dir, "git", "commit", "-m", "add gitignore")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("secret data"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "secret"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "secret", "key.pem"), []byte("key"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("tracked"), 0644))

	return dir
}

func TestListIgnoredFiles(t *testing.T) {
	dir := setupRepoWithIgnored(t)
	wt := git.NewExecWorktree(dir)

	files, err := wt.ListIgnoredFiles()
	require.NoError(t, err)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}

	assert.Contains(t, paths, "ignored.txt")
	assert.Contains(t, paths, "secret/")
	assert.NotContains(t, paths, "tracked.txt")
	assert.NotContains(t, paths, ".gitignore")
}

func TestListIgnoredFiles_NoIgnored(t *testing.T) {
	dir := setupTestRepo(t)
	wt := git.NewExecWorktree(dir)

	files, err := wt.ListIgnoredFiles()
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestIsIgnored_IgnoredFile(t *testing.T) {
	dir := setupRepoWithIgnored(t)
	wt := git.NewExecWorktree(dir)

	ignored, err := wt.IsIgnored("ignored.txt")
	require.NoError(t, err)
	assert.True(t, ignored)
}

func TestIsIgnored_TrackedFile(t *testing.T) {
	dir := setupRepoWithIgnored(t)
	wt := git.NewExecWorktree(dir)

	ignored, err := wt.IsIgnored("tracked.txt")
	require.NoError(t, err)
	assert.False(t, ignored)
}

func TestIsIgnored_Directory(t *testing.T) {
	dir := setupRepoWithIgnored(t)
	wt := git.NewExecWorktree(dir)

	ignored, err := wt.IsIgnored("secret/")
	require.NoError(t, err)
	assert.True(t, ignored)
}
