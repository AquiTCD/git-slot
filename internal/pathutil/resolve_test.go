package pathutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
)

func homeDir(t *testing.T) string {
	t.Helper()
	h, err := os.UserHomeDir()
	require.NoError(t, err)
	return h
}

func TestResolveSlotsBasePath_DefaultGwqStyle(t *testing.T) {
	home := homeDir(t)
	cfg := &config.Config{}
	remote := &git.RemoteInfo{Host: "github.com", Owner: "user", Repo: "repo"}

	got, err := ResolveSlotsBasePath(cfg, remote)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "worktrees", "github.com", "user", "repo", "slots"), got)
}

func TestResolveSlotsBasePath_CustomWtBasePath_Absolute(t *testing.T) {
	cfg := &config.Config{WtBasePath: "/opt/worktrees"}
	remote := &git.RemoteInfo{Host: "gitlab.com", Owner: "team", Repo: "project"}

	got, err := ResolveSlotsBasePath(cfg, remote)
	require.NoError(t, err)
	assert.Equal(t, "/opt/worktrees/gitlab.com/team/project/slots", got)
}

func TestResolveSlotsBasePath_CustomWtBasePath_Tilde(t *testing.T) {
	home := homeDir(t)
	cfg := &config.Config{WtBasePath: "~/custom-trees"}
	remote := &git.RemoteInfo{Host: "github.com", Owner: "org", Repo: "app"}

	got, err := ResolveSlotsBasePath(cfg, remote)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "custom-trees", "github.com", "org", "app", "slots"), got)
}

func TestResolveSlotsBasePath_NoRemoteInfo(t *testing.T) {
	cfg := &config.Config{}
	_, err := ResolveSlotsBasePath(cfg, nil)
	require.ErrorIs(t, err, ErrNoRemoteInfo)
}

func TestResolveSlotsBasePath_NoRemoteInfo_WithWtBasePath(t *testing.T) {
	cfg := &config.Config{WtBasePath: "/some/base"}
	_, err := ResolveSlotsBasePath(cfg, nil)
	require.ErrorIs(t, err, ErrNoRemoteInfo)
}

func TestResolveSlotPath(t *testing.T) {
	got := ResolveSlotPath("/base/path/slots", "dev")
	assert.Equal(t, filepath.Join("/base/path/slots", "dev"), got)
}

func TestExpandHome_TildeSlashPrefix(t *testing.T) {
	home := homeDir(t)
	got, err := ExpandHome("~/foo")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "foo"), got)
}

func TestExpandHome_AbsolutePath(t *testing.T) {
	got, err := ExpandHome("/absolute/path")
	require.NoError(t, err)
	assert.Equal(t, "/absolute/path", got)
}

func TestExpandHome_RelativePath(t *testing.T) {
	got, err := ExpandHome("relative/path")
	require.NoError(t, err)
	assert.Equal(t, "relative/path", got)
}

func TestExpandHome_TildeOnly(t *testing.T) {
	home := homeDir(t)
	got, err := ExpandHome("~")
	require.NoError(t, err)
	assert.Equal(t, home, got)
}
