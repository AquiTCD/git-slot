package slot

import (
	"errors"
	"fmt"
)

var (
	ErrSlotUnknown      = errors.New("slot is not defined in configuration")
	ErrSlotAlreadyEmpty = errors.New("slot is already empty")
	ErrSlotDirty        = errors.New("slot has uncommitted changes")
	ErrSlotEmpty        = errors.New("slot is empty")
	ErrBranchNotFound   = errors.New("branch not found")
	ErrBranchInUse      = errors.New("branch is already in use")
	ErrBranchExists     = errors.New("branch already exists")
)

type SlotError struct {
	SlotName string
	Err      error
	Detail   string
}

func (e *SlotError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("slot '%s': %s: %s", e.SlotName, e.Err, e.Detail)
	}
	return fmt.Sprintf("slot '%s': %s", e.SlotName, e.Err)
}

func (e *SlotError) Unwrap() error {
	return e.Err
}

type BranchError struct {
	Branch string
	UsedBy string
	Err    error
}

func (e *BranchError) Error() string {
	if e.UsedBy != "" {
		return fmt.Sprintf("branch '%s': %s (used by: %s)", e.Branch, e.Err, e.UsedBy)
	}
	return fmt.Sprintf("branch '%s': %s", e.Branch, e.Err)
}

func (e *BranchError) Unwrap() error {
	return e.Err
}
