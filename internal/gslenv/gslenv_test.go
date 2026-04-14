package gslenv

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripFromEnv_RemovesWrapperVar(t *testing.T) {
	in := []string{"FOO=bar", "GSL_FROM_WRAPPER=1", "GSL_SLOT_NAME=work"}
	got := StripFromEnv(in)
	assert.ElementsMatch(t, []string{"FOO=bar", "GSL_SLOT_NAME=work"}, got)
}

func TestFromWrapper_TrueOnlyForValueOne(t *testing.T) {
	t.Setenv(Name, "1")
	assert.True(t, FromWrapper())
}

func TestFromWrapper_FalseWhenUnset(t *testing.T) {
	t.Setenv(Name, "")
	assert.False(t, FromWrapper())
}

func TestFromWrapper_FalseForOtherValues(t *testing.T) {
	t.Setenv(Name, "yes")
	assert.False(t, FromWrapper())
}

func TestStripFromOSEnviron(t *testing.T) {
	t.Setenv(Name, "1")
	t.Setenv("GSLENV_TEST_KEEP", "x")
	got := StripFromOSEnviron()
	joined := strings.Join(got, "\n")
	assert.NotContains(t, joined, Name+"=")
	assert.Contains(t, joined, "GSLENV_TEST_KEEP=x")
}
