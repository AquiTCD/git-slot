package hook

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte("#!/bin/sh\n"+content), 0o755)
	require.NoError(t, err)
	return path
}

func TestRun_EmptyActions(t *testing.T) {
	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run(nil, HookEnv{})
	assert.NoError(t, err)
}

func TestRun_CommandNotFound(t *testing.T) {
	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{Type: "run", Command: "/nonexistent/hook.sh"}}, HookEnv{})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHookFailed)
}

func TestRun_NotExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits not enforced on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "hook.sh")
	err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o644)
	require.NoError(t, err)

	r := NewRunner(os.Stdout, os.Stderr)
	err = r.Run([]config.HookAction{{Type: "run", Command: path}}, HookEnv{})

	require.Error(t, err)
	// Now it defaults to running as sh -c path, so it might fail with different error if not executable
	assert.ErrorIs(t, err, ErrHookFailed)
}

func TestRun_Success(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "ok.sh", "exit 0\n")

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{Type: "run", Command: path}}, HookEnv{SlotName: "work", Action: "mount"})

	assert.NoError(t, err)
}

func TestRun_FailedExitCode(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "fail.sh", "exit 1\n")

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{Type: "run", Command: path}}, HookEnv{})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHookFailed)
	assert.Contains(t, err.Error(), "exit code 1")
}

func TestRun_EnvVarsSet(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "env.sh", `
echo "SLOT_NAME=$GSL_SLOT_NAME"
echo "SLOT_PATH=$GSL_SLOT_PATH"
echo "BRANCH=$GSL_BRANCH"
echo "REPO_ROOT=$GSL_REPO_ROOT"
echo "ACTION=$GSL_ACTION"
`)

	var stdout bytes.Buffer
	r := NewRunner(&stdout, os.Stderr)
	env := HookEnv{
		SlotName: "work",
		SlotPath: "/slots/work",
		Branch:   "feature/x",
		RepoRoot: "/repo",
		Action:   "mount",
	}
	err := r.Run([]config.HookAction{{Type: "run", Command: path}}, env)

	require.NoError(t, err)
	out := stdout.String()
	assert.Contains(t, out, "SLOT_NAME=work")
	assert.Contains(t, out, "SLOT_PATH=/slots/work")
	assert.Contains(t, out, "BRANCH=feature/x")
	assert.Contains(t, out, "REPO_ROOT=/repo")
	assert.Contains(t, out, "ACTION=mount")
}

func TestRun_Timeout(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "slow.sh", "sleep 10\n")

	r := NewRunner(os.Stdout, os.Stderr)
	r.timeout = 100 * time.Millisecond
	err := r.Run([]config.HookAction{{Type: "run", Command: path}}, HookEnv{})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHookTimeout)
}

func TestRun_StdoutStderr(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "output.sh", `
echo "hello stdout"
echo "hello stderr" >&2
`)

	var stdout, stderr bytes.Buffer
	r := NewRunner(&stdout, &stderr)
	err := r.Run([]config.HookAction{{Type: "run", Command: path}}, HookEnv{})

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "hello stdout")
	assert.Contains(t, stderr.String(), "hello stderr")
}
