package hook

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestHandleLink_CreatesSymlink(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("hello"), 0o644))

	dest := filepath.Join(dir, "sub", "link.txt")

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "link",
		Source: src,
		Dest:   dest,
	}}, HookEnv{})

	require.NoError(t, err)

	target, err := os.Readlink(dest)
	require.NoError(t, err)
	assert.Equal(t, src, target)
}

func TestHandleLink_ReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("new"), 0o644))

	dest := filepath.Join(dir, "link.txt")
	require.NoError(t, os.WriteFile(dest, []byte("old"), 0o644))

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "link",
		Source: src,
		Dest:   dest,
	}}, HookEnv{})

	require.NoError(t, err)

	target, err := os.Readlink(dest)
	require.NoError(t, err)
	assert.Equal(t, src, target)
}

func TestHandleLink_MissingSrc(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "link.txt")

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "link",
		Source: "",
		Dest:   dest,
	}}, HookEnv{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing source or destination")
}

func TestHandleLink_ExpandsEnvVars(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0o644))

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "link",
		Source: filepath.Join("$GSL_REPO_ROOT", "source.txt"),
		Dest:   filepath.Join("$GSL_SLOT_PATH", "source.txt"),
	}}, HookEnv{
		RepoRoot: dir,
		SlotPath: filepath.Join(dir, "slot"),
	})

	require.NoError(t, err)

	dest := filepath.Join(dir, "slot", "source.txt")
	target, err := os.Readlink(dest)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "source.txt"), target)
}

func TestHandleCopy_CopiesFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("content"), 0o644))

	dest := filepath.Join(dir, "sub", "copy.txt")

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "copy",
		Source: src,
		Dest:   dest,
	}}, HookEnv{})

	require.NoError(t, err)

	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))
}

func TestHandleCopy_CopiesDirectory(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "srcdir")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0o644))

	dest := filepath.Join(dir, "destdir")

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "copy",
		Source: srcDir,
		Dest:   dest,
	}}, HookEnv{})

	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dest, "a.txt"))
	require.NoError(t, err)
	assert.Equal(t, "aaa", string(data))
}

func TestHandleCopy_MissingSrc(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "copy.txt")

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "copy",
		Source: "",
		Dest:   dest,
	}}, HookEnv{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing source or destination")
}

func TestHandleCopy_ReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("new"), 0o644))

	dest := filepath.Join(dir, "copy.txt")
	require.NoError(t, os.WriteFile(dest, []byte("old"), 0o644))

	r := NewRunner(os.Stdout, os.Stderr)
	err := r.Run([]config.HookAction{{
		Type:   "copy",
		Source: src,
		Dest:   dest,
	}}, HookEnv{})

	require.NoError(t, err)

	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "new", string(data))
}

func TestOsEnvironWithoutGslWrapper(t *testing.T) {
	t.Setenv("GSL_FROM_WRAPPER", "1")
	t.Setenv("HOOK_TEST_KEEP", "yes")
	got := osEnvironWithoutGslWrapper()
	joined := strings.Join(got, "\n")
	assert.NotContains(t, joined, "GSL_FROM_WRAPPER=")
	assert.Contains(t, joined, "HOOK_TEST_KEEP=yes")
}
