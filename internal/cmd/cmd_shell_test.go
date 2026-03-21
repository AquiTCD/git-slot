package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

func TestCheckShellNesting(t *testing.T) {
	t.Run("no nesting when GSL_SHELL_SESSION is unset", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "")
		err := checkShellNesting()
		assert.NoError(t, err)
	})

	t.Run("nesting detected when GSL_SHELL_SESSION is set", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		err := checkShellNesting()
		assert.ErrorIs(t, err, ErrShellNested)
	})
}

func TestCheckShellNestingForSet(t *testing.T) {
	t.Run("no nesting when GSL_SHELL_SESSION is unset", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "")
		err := checkShellNestingForSet("any-slot")
		assert.NoError(t, err)
	})

	t.Run("same slot is allowed", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "main-work")
		err := checkShellNestingForSet("main-work")
		assert.NoError(t, err)
	})

	t.Run("different slot is blocked", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "main-work")
		err := checkShellNestingForSet("hotfix")
		assert.ErrorIs(t, err, ErrShellNested)
	})
}

func TestRunMount_GetPathError(t *testing.T) {
	t.Run("returns error when GetPath fails after mount", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "")

		mgr := &mockSlotManager{
			mountFn:   func(_, _ string, _ slot.MountOptions) error { return nil },
			getPathFn: func(_ string) (string, error) { return "", errors.New("worktree not found") },
		}
		a := &app{
			mgr:      mgr,
			cfg:      &config.Config{LaunchShell: nil},
			basePath: "/slots",
			repoRoot: "/repo",
		}

		var buf bytes.Buffer
		err := runMount(a, "work", "feat/x", mountOptions{}, &buf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "worktree not found")
		assert.Empty(t, buf.String())
	})
}

func TestRunMount_LaunchShell(t *testing.T) {
	var shellLaunched bool
	origExecShell := execShellFunc
	execShellFunc = func(dir string, env []string) error {
		shellLaunched = true
		return nil
	}
	t.Cleanup(func() { execShellFunc = origExecShell })

	t.Run("launch_shell=true launches shell after mount", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "")
		shellLaunched = false

		mgr := &mockSlotManager{
			mountFn: func(_, _ string, _ slot.MountOptions) error { return nil },
			statusFn: func(name string) (*slot.SlotStatus, error) {
				return &slot.SlotStatus{
					Slot: slot.Slot{
						Name:   name,
						Branch: "feat/x",
						State:  slot.SlotActive,
						Path:   "/slots/" + name,
					},
				}, nil
			},
		}

		a := &app{
			mgr:      mgr,
			cfg:      &config.Config{LaunchShell: boolPtr(true), Slots: []config.SlotDefinition{{Name: "work"}}},
			basePath: "/slots",
			repoRoot: "/repo",
		}

		var buf bytes.Buffer
		err := runMount(a, "work", "feat/x", mountOptions{}, &buf)
		require.NoError(t, err)
		assert.True(t, shellLaunched)
	})

	t.Run("launch_shell=true with --no-shell outputs path", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "")
		shellLaunched = false

		mgr := &mockSlotManager{
			mountFn:   func(_, _ string, _ slot.MountOptions) error { return nil },
			getPathFn: func(name string) (string, error) { return "/slots/" + name, nil },
		}

		a := &app{
			mgr:      mgr,
			cfg:      &config.Config{LaunchShell: boolPtr(true)},
			basePath: "/slots",
			repoRoot: "/repo",
		}

		var buf bytes.Buffer
		err := runMount(a, "work", "feat/x", mountOptions{noShell: true}, &buf)
		require.NoError(t, err)
		assert.False(t, shellLaunched)
		assert.Contains(t, buf.String(), "/slots/work")
	})

	t.Run("launch_shell=true blocks different slot in shell session", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "main-work")
		shellLaunched = false

		mgr := &mockSlotManager{}
		a := &app{
			mgr:      mgr,
			cfg:      &config.Config{LaunchShell: boolPtr(true)},
			basePath: "/slots",
			repoRoot: "/repo",
		}

		var buf bytes.Buffer
		err := runMount(a, "hotfix", "fix/bug", mountOptions{}, &buf)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrShellNested)
		assert.False(t, shellLaunched)
	})

	t.Run("launch_shell=true allows same slot switch in shell session", func(t *testing.T) {
		t.Setenv("GSL_SHELL_SESSION", "1")
		t.Setenv("GSL_SLOT_NAME", "main-work")
		shellLaunched = false

		mgr := &mockSlotManager{
			mountFn:   func(_, _ string, _ slot.MountOptions) error { return nil },
			getPathFn: func(name string) (string, error) { return "/slots/" + name, nil },
		}

		a := &app{
			mgr:      mgr,
			cfg:      &config.Config{LaunchShell: boolPtr(true)},
			basePath: "/slots",
			repoRoot: "/repo",
		}

		var buf bytes.Buffer
		err := runMount(a, "main-work", "other-branch", mountOptions{}, &buf)
		require.NoError(t, err)
		assert.False(t, shellLaunched, "same-slot switch should not launch a new shell")
		assert.Contains(t, buf.String(), "/slots/main-work")
	})
}
