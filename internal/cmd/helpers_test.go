package cmd

import (
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestStripWrapperExclusiveEnv(t *testing.T) {
	in := []string{"FOO=bar", "GSL_FROM_WRAPPER=1", "GSL_SLOT_NAME=work"}
	got := stripWrapperExclusiveEnv(in)
	assert.ElementsMatch(t, []string{"FOO=bar", "GSL_SLOT_NAME=work"}, got)
}

func TestWantShell_GslWrapperSuppressesEvenWhenLaunchShellTrue(t *testing.T) {
	t.Setenv(envGslFromWrapper, "1")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.False(t, wantShell(cfg, false))
}

func TestWantShell_RespectsNoShellFlag(t *testing.T) {
	t.Setenv(envGslFromWrapper, "")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.False(t, wantShell(cfg, true))
}

func TestWantShell_LaunchShellTrueWithoutWrapper(t *testing.T) {
	t.Setenv(envGslFromWrapper, "")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.True(t, wantShell(cfg, false))
}

func TestWantShell_GslWrapperValueNotOneDoesNotSuppress(t *testing.T) {
	t.Setenv(envGslFromWrapper, "yes")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.True(t, wantShell(cfg, false))
}
