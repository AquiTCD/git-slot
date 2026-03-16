package errutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExitError(t *testing.T) {
	err := NewExitError("something failed", 42)

	assert.EqualError(t, err, "something failed")
	assert.Equal(t, 42, err.ExitCode())
}

func TestExitError_ImplementsErrorInterface(t *testing.T) {
	var err error = NewExitError("test", 1)
	assert.EqualError(t, err, "test")
}

func TestExitError_ExitCodeZero(t *testing.T) {
	err := NewExitError("ok", 0)
	assert.Equal(t, 0, err.ExitCode())
}

func TestExitError_SatisfiesInterface(t *testing.T) {
	err := NewExitError("msg", 2)
	ex := err
	assert.Equal(t, 2, ex.ExitCode())
	assert.EqualError(t, ex, "msg")
}
