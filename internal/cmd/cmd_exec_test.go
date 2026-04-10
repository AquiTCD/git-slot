package cmd

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/errutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunExecFromApp_Success(t *testing.T) {
	orig := execRunHook
	t.Cleanup(func() { execRunHook = orig })

	var got *exec.Cmd
	execRunHook = func(c *exec.Cmd) error {
		got = c
		return nil
	}

	mgr := &mockSlotManager{
		statusFn: func(name string) (*slot.SlotStatus, error) {
			return &slot.SlotStatus{
				Slot: slot.Slot{
					Name:   name,
					State:  slot.SlotActive,
					Branch: "main",
					Path:   "/wt/r@work",
				},
			}, nil
		},
	}
	a := &app{
		mgr:      mgr,
		cfg:      &config.Config{Slots: []config.SlotDefinition{{Name: "work"}}},
		basePath: "/wt",
		repoName: "r",
		repoRoot: "/repo/main",
	}

	err := runExecFromApp(a, []string{"git-slot", "exec", "work", "--", "echo", "ok"})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/wt/r@work", got.Dir)
	require.GreaterOrEqual(t, len(got.Args), 2)
	assert.Equal(t, "ok", got.Args[len(got.Args)-1])
	assert.Contains(t, got.Env, "GSL_SLOT_NAME=work")
	foundSession := false
	for _, e := range got.Env {
		if e == "GSL_SHELL_SESSION=1" {
			foundSession = true
			break
		}
	}
	assert.False(t, foundSession)
}

func TestRunExecFromApp_PropagatesExitCode(t *testing.T) {
	dir := t.TempDir()
	orig := execRunHook
	t.Cleanup(func() { execRunHook = orig })

	execRunHook = func(c *exec.Cmd) error {
		return c.Run()
	}

	mgr := &mockSlotManager{
		statusFn: func(name string) (*slot.SlotStatus, error) {
			return &slot.SlotStatus{
				Slot: slot.Slot{Name: name, State: slot.SlotActive, Branch: "b", Path: dir},
			}, nil
		},
	}
	a := &app{
		mgr:      mgr,
		cfg:      &config.Config{Slots: []config.SlotDefinition{{Name: "work"}}},
		basePath: "/wt",
		repoName: "r",
		repoRoot: "/repo",
	}

	err := runExecFromApp(a, []string{"git-slot", "exec", "work", "--", "sh", "-c", "exit 42"})
	require.Error(t, err)
	assert.Equal(t, 42, mapExitCode(err))
	var ex errutil.ExitError
	require.True(t, errors.As(err, &ex))
	assert.Equal(t, 42, ex.ExitCode())
}
