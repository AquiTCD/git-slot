package slot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlot_DisplayState_Empty(t *testing.T) {
	s := &Slot{State: SlotEmpty}
	assert.Equal(t, "empty", s.DisplayState())
}

func TestSlot_DisplayState_Active(t *testing.T) {
	s := &Slot{State: SlotActive, IsDirty: false}
	assert.Equal(t, "active", s.DisplayState())
}

func TestSlot_DisplayState_Dirty(t *testing.T) {
	s := &Slot{State: SlotActive, IsDirty: true}
	assert.Equal(t, "dirty", s.DisplayState())
}

func TestSlotError_Unwrap(t *testing.T) {
	err := &SlotError{SlotName: "dev", Err: ErrSlotDirty}
	require.True(t, errors.Is(err, ErrSlotDirty))
}

func TestSlotError_Message(t *testing.T) {
	err := &SlotError{SlotName: "dev", Err: ErrSlotDirty}
	assert.Contains(t, err.Error(), "dev")
	assert.Contains(t, err.Error(), ErrSlotDirty.Error())
}

func TestSlotError_WithDetail(t *testing.T) {
	err := &SlotError{SlotName: "dev", Err: ErrSlotDirty, Detail: "3 files modified"}
	msg := err.Error()
	assert.Contains(t, msg, "dev")
	assert.Contains(t, msg, "3 files modified")
}

func TestBranchError_Unwrap(t *testing.T) {
	err := &BranchError{Branch: "feature/login", Err: ErrBranchInUse, UsedBy: "slot-1"}
	require.True(t, errors.Is(err, ErrBranchInUse))
}

func TestBranchError_Message(t *testing.T) {
	err := &BranchError{Branch: "feature/login", Err: ErrBranchInUse, UsedBy: "slot-1"}
	msg := err.Error()
	assert.Contains(t, msg, "feature/login")
	assert.Contains(t, msg, "slot-1")
}

func TestBranchError_WithoutUsedBy(t *testing.T) {
	err := &BranchError{Branch: "feature/login", Err: ErrBranchNotFound}
	msg := err.Error()
	assert.Contains(t, msg, "feature/login")
	assert.NotContains(t, msg, "used by")
}
