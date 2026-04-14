package cmd

import (
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/gslenv"
	"github.com/stretchr/testify/assert"
)

func TestWantShell_GslWrapperSuppressesEvenWhenLaunchShellTrue(t *testing.T) {
	t.Setenv(gslenv.Name, "1")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.False(t, wantShell(cfg, false))
}

func TestWantShell_RespectsNoShellFlag(t *testing.T) {
	t.Setenv(gslenv.Name, "")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.False(t, wantShell(cfg, true))
}

func TestWantShell_LaunchShellTrueWithoutWrapper(t *testing.T) {
	t.Setenv(gslenv.Name, "")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.True(t, wantShell(cfg, false))
}

func TestWantShell_GslWrapperValueNotOneDoesNotSuppress(t *testing.T) {
	t.Setenv(gslenv.Name, "yes")
	on := true
	cfg := &config.Config{LaunchShell: &on}
	assert.True(t, wantShell(cfg, false))
}
