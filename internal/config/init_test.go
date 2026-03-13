package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_ProjectConfig(t *testing.T) {
	dir := t.TempDir()

	path, err := Init(InitOptions{}, dir)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "git-slot.toml"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "[[slots]]")
	assert.Contains(t, string(data), `name = "slot-1"`)
}

func TestInit_GlobalConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := Init(InitOptions{Global: true}, "")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "git-slot", "config.toml"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "[[slots]]")
}

func TestInit_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "git-slot.toml")
	require.NoError(t, os.WriteFile(existing, []byte("existing"), 0o644))

	_, err := Init(InitOptions{}, dir)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigExists)

	data, err := os.ReadFile(existing)
	require.NoError(t, err)
	assert.Equal(t, "existing", string(data))
}

func TestInit_AlreadyExists_Force(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "git-slot.toml")
	require.NoError(t, os.WriteFile(existing, []byte("old"), 0o644))

	path, err := Init(InitOptions{Force: true}, dir)

	require.NoError(t, err)
	assert.Equal(t, existing, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "[[slots]]")
}

func TestInit_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "nested", "repo")

	path, err := Init(InitOptions{}, nested)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(nested, "git-slot.toml"), path)

	_, err = os.Stat(path)
	require.NoError(t, err)
}
