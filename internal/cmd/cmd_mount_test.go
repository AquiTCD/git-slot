package cmd

import (
	"bytes"
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleInteractiveResult_ActiveSlotWithLaunchShell(t *testing.T) {
	t.Setenv("GSL_SHELL_SESSION", "")

	var shellLaunched bool
	origExecShell := execShellFunc
	execShellFunc = func(dir string, env []string) error {
		shellLaunched = true
		return nil
	}
	t.Cleanup(func() { execShellFunc = origExecShell })

	mgr := &mockSlotManager{
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
		mgr: mgr,
		cfg: &config.Config{LaunchShell: boolPtr(true)},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := handleInteractiveResult(a, tui.Result{SlotName: "wood"}, false, false, &buf)
	require.NoError(t, err)
	assert.True(t, shellLaunched)
	assert.Empty(t, buf.String())
}

func TestHandleInteractiveResult_ActiveSlotWithoutLaunchShell(t *testing.T) {
	mgr := &mockSlotManager{
		getPathFn: func(name string) (string, error) {
			return "/slots/" + name, nil
		},
	}
	a := &app{
		mgr: mgr,
		cfg: &config.Config{LaunchShell: boolPtr(false)},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := handleInteractiveResult(a, tui.Result{SlotName: "wood"}, false, false, &buf)
	require.NoError(t, err)
	assert.Equal(t, "/slots/wood\n", buf.String())
}

func TestHandleInteractiveResult_NestedShellDifferentSlot(t *testing.T) {
	t.Setenv("GSL_SHELL_SESSION", "1")
	t.Setenv("GSL_SLOT_NAME", "fire")

	a := &app{
		mgr: &mockSlotManager{},
		cfg: &config.Config{LaunchShell: boolPtr(true)},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := handleInteractiveResult(a, tui.Result{SlotName: "wood"}, false, false, &buf)
	require.ErrorIs(t, err, ErrShellNested)
}

func TestRunGetPathOrLaunchShell_SameSlotInsideShellPrintsPath(t *testing.T) {
	t.Setenv("GSL_SHELL_SESSION", "1")
	t.Setenv("GSL_SLOT_NAME", "wood")

	var shellLaunched bool
	origExecShell := execShellFunc
	execShellFunc = func(dir string, env []string) error {
		shellLaunched = true
		return nil
	}
	t.Cleanup(func() { execShellFunc = origExecShell })

	mgr := &mockSlotManager{
		getPathFn: func(name string) (string, error) {
			return "/slots/" + name, nil
		},
		statusFn: func(name string) (*slot.SlotStatus, error) {
			return &slot.SlotStatus{
				Slot: slot.Slot{Name: name, State: slot.SlotActive, Path: "/slots/" + name},
			}, nil
		},
	}
	a := &app{
		mgr: mgr,
		cfg: &config.Config{LaunchShell: boolPtr(true)},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runGetPathOrLaunchShell(a, "wood", false, &buf)
	require.NoError(t, err)
	assert.False(t, shellLaunched)
	assert.Equal(t, "/slots/wood\n", buf.String())
}

func TestRunGetPathOrLaunchShell_EmptySlotPrintsPath(t *testing.T) {
	t.Setenv("GSL_SHELL_SESSION", "")

	var shellLaunched bool
	origExecShell := execShellFunc
	execShellFunc = func(dir string, env []string) error {
		shellLaunched = true
		return nil
	}
	t.Cleanup(func() { execShellFunc = origExecShell })

	mgr := &mockSlotManager{
		getPathFn: func(name string) (string, error) {
			return "/reserved/" + name, nil
		},
		statusFn: func(name string) (*slot.SlotStatus, error) {
			return &slot.SlotStatus{Slot: slot.Slot{Name: name, State: slot.SlotEmpty}}, nil
		},
	}
	a := &app{
		mgr: mgr,
		cfg: &config.Config{LaunchShell: boolPtr(true)},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runGetPathOrLaunchShell(a, "idle", false, &buf)
	require.NoError(t, err)
	assert.False(t, shellLaunched)
	assert.Equal(t, "/reserved/idle\n", buf.String())
}
