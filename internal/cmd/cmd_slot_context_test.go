package cmd

import (
	"errors"
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/errutil"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadActiveSlotContext(t *testing.T) {
	mgr := &mockSlotManager{
		statusFn: func(name string) (*slot.SlotStatus, error) {
			if name == "empty" {
				return &slot.SlotStatus{Slot: slot.Slot{Name: name, State: slot.SlotEmpty, Path: "/x"}}, nil
			}
			return &slot.SlotStatus{
				Slot: slot.Slot{Name: name, State: slot.SlotActive, Branch: "b", Path: "/p/" + name},
			}, nil
		},
	}
	a := &app{
		mgr: mgr,
		cfg: &config.Config{
			Slots: []config.SlotDefinition{
				{Name: "work", Env: map[string]string{"PORT": "1"}},
			},
		},
		repoRoot: "/repo",
	}

	t.Run("active", func(t *testing.T) {
		st, info, env, err := loadActiveSlotContext(a, "work")
		require.NoError(t, err)
		assert.Equal(t, slot.SlotActive, st.State)
		assert.Equal(t, "work", info.SlotName)
		assert.Equal(t, "1", env["PORT"])
	})

	t.Run("empty slot", func(t *testing.T) {
		_, _, _, err := loadActiveSlotContext(a, "empty")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})
}

func TestExitIfNotSlotWorktree(t *testing.T) {
	err := exitIfNotSlotWorktree(slot.ErrNotASlotWorktree, "nope")
	var ex errutil.ExitError
	require.True(t, errors.As(err, &ex))
	assert.Equal(t, 1, ex.ExitCode())
	assert.Contains(t, err.Error(), "nope")

	assert.Equal(t, errors.New("x"), exitIfNotSlotWorktree(errors.New("x"), "ignored"))
}
