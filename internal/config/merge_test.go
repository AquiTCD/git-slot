package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge_ScalarOverride(t *testing.T) {
	tests := []struct {
		name     string
		base     *Config
		override *Config
		wantDir  string
		wantPath string
	}{
		{
			name:     "override gwq_basedir wins",
			base:     &Config{GwqBaseDir: "/base"},
			override: &Config{GwqBaseDir: "/override"},
			wantDir:  "/override",
		},
		{
			name:     "empty override preserves base gwq_basedir",
			base:     &Config{GwqBaseDir: "/base"},
			override: &Config{},
			wantDir:  "/base",
		},
		{
			name:     "override slots_base_path wins",
			base:     &Config{SlotsBasePath: "/base/slots"},
			override: &Config{SlotsBasePath: "/override/slots"},
			wantPath: "/override/slots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.override)
			if tt.wantDir != "" {
				assert.Equal(t, tt.wantDir, got.GwqBaseDir)
			}
			if tt.wantPath != "" {
				assert.Equal(t, tt.wantPath, got.SlotsBasePath)
			}
		})
	}
}

func TestMerge_SlotsReplacement(t *testing.T) {
	base := &Config{Slots: slots("dev", "staging")}

	t.Run("non-nil override replaces base slots", func(t *testing.T) {
		override := &Config{Slots: slots("prod")}
		got := Merge(base, override)

		require.Len(t, got.Slots, 1)
		assert.Equal(t, "prod", got.Slots[0].Name)
	})

	t.Run("empty override slots replaces base slots", func(t *testing.T) {
		override := &Config{Slots: []SlotDefinition{}}
		got := Merge(base, override)

		require.NotNil(t, got.Slots)
		assert.Empty(t, got.Slots)
	})

	t.Run("nil override slots preserves base slots", func(t *testing.T) {
		override := &Config{Slots: nil}
		got := Merge(base, override)

		require.Len(t, got.Slots, 2)
		assert.Equal(t, "dev", got.Slots[0].Name)
		assert.Equal(t, "staging", got.Slots[1].Name)
	})
}

func TestMerge_HooksFieldLevel(t *testing.T) {
	t.Run("disjoint hooks are both preserved", func(t *testing.T) {
		base := &Config{Hooks: HooksConfig{PreClear: "echo pre"}}
		override := &Config{Hooks: HooksConfig{PostLoad: "echo post"}}

		got := Merge(base, override)

		assert.Equal(t, "echo pre", got.Hooks.PreClear)
		assert.Equal(t, "echo post", got.Hooks.PostLoad)
	})

	t.Run("same hook field override wins", func(t *testing.T) {
		base := &Config{Hooks: HooksConfig{PostLoad: "base-hook"}}
		override := &Config{Hooks: HooksConfig{PostLoad: "override-hook"}}

		got := Merge(base, override)

		assert.Equal(t, "override-hook", got.Hooks.PostLoad)
	})
}

func TestMerge_BothEmpty(t *testing.T) {
	got := Merge(&Config{}, &Config{})

	assert.Empty(t, got.GwqBaseDir)
	assert.Empty(t, got.SlotsBasePath)
	assert.Nil(t, got.Slots)
	assert.Empty(t, got.Hooks)
}

func TestMerge_NilBase(t *testing.T) {
	override := &Config{
		GwqBaseDir: "/from-override",
		Slots:      slots("alpha"),
	}

	got := Merge(nil, override)

	require.NotNil(t, got)
	assert.Equal(t, "/from-override", got.GwqBaseDir)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, "alpha", got.Slots[0].Name)
}

func TestMerge_NilOverride(t *testing.T) {
	base := &Config{
		GwqBaseDir: "/from-base",
		Slots:      slots("beta"),
	}

	got := Merge(base, nil)

	require.NotNil(t, got)
	assert.Equal(t, "/from-base", got.GwqBaseDir)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, "beta", got.Slots[0].Name)
}

func TestMerge_BothNil(t *testing.T) {
	got := Merge(nil, nil)

	require.NotNil(t, got)
	assert.Empty(t, got.GwqBaseDir)
	assert.Nil(t, got.Slots)
}

func TestMerge_FullIntegration(t *testing.T) {
	base := &Config{
		GwqBaseDir:    "/home/user/gwq",
		SlotsBasePath: "/home/user/slots",
		Slots:         slots("dev", "staging", "prod"),
		Hooks: HooksConfig{
			PreLoad:   "echo pre-load",
			PostLoad:  "echo post-load",
			PreClear:  "echo pre-clear",
			PostClear: "echo post-clear",
		},
	}
	override := &Config{
		SlotsBasePath: "/custom/slots",
		Slots:         slots("only-one"),
		Hooks: HooksConfig{
			PostLoad: "custom-post-load",
		},
	}

	got := Merge(base, override)

	assert.Equal(t, "/home/user/gwq", got.GwqBaseDir)
	assert.Equal(t, "/custom/slots", got.SlotsBasePath)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, "only-one", got.Slots[0].Name)
	assert.Equal(t, "echo pre-load", got.Hooks.PreLoad)
	assert.Equal(t, "custom-post-load", got.Hooks.PostLoad)
	assert.Equal(t, "echo pre-clear", got.Hooks.PreClear)
	assert.Equal(t, "echo post-clear", got.Hooks.PostClear)
}

func TestMerge_DoesNotMutateInputs(t *testing.T) {
	base := &Config{
		GwqBaseDir: "/base",
		Slots:      slots("a", "b"),
		Hooks:      HooksConfig{PreLoad: "base-hook"},
	}
	override := &Config{
		GwqBaseDir: "/override",
		Slots:      slots("x"),
	}

	got := Merge(base, override)

	assert.Equal(t, "/base", base.GwqBaseDir)
	assert.Equal(t, "/override", override.GwqBaseDir)
	require.Len(t, base.Slots, 2)
	require.Len(t, override.Slots, 1)

	got.Slots[0].Name = "mutated"
	assert.Equal(t, "x", override.Slots[0].Name)
	assert.Equal(t, "a", base.Slots[0].Name)
}
