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
		wantPath string
	}{
		{
			name:     "override wt_base_path wins",
			base:     &Config{WtBasePath: "/base/slots"},
			override: &Config{WtBasePath: "/override/slots"},
			wantPath: "/override/slots",
		},
		{
			name:     "empty override preserves base wt_base_path",
			base:     &Config{WtBasePath: "/base/slots"},
			override: &Config{},
			wantPath: "/base/slots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.override)
			if tt.wantPath != "" {
				assert.Equal(t, tt.wantPath, got.WtBasePath)
			}
		})
	}
}

func TestMerge_SlotsMerging(t *testing.T) {
	base := &Config{Slots: []SlotDefinition{{Name: "dev"}, {Name: "staging"}}}

	t.Run("disjoint slots are merged", func(t *testing.T) {
		override := &Config{Slots: []SlotDefinition{{Name: "prod"}}}
		got := Merge(base, override)

		require.Len(t, got.Slots, 3)
		assert.ElementsMatch(t, []string{"dev", "staging", "prod"}, slotNames(got.Slots))
	})

	t.Run("existing slot is overwritten by name", func(t *testing.T) {
		override := &Config{Slots: []SlotDefinition{{Name: "dev", Icon: "🚀"}}}
		got := Merge(base, override)

		require.Len(t, got.Slots, 2)
		assert.Equal(t, "🚀", findSlot(got.Slots, "dev").Icon)
	})

	t.Run("nil override slots preserves base slots", func(t *testing.T) {
		override := &Config{Slots: nil}
		got := Merge(base, override)

		require.Len(t, got.Slots, 2)
	})
}

func slotNames(slots []SlotDefinition) []string {
	names := make([]string, len(slots))
	for i, s := range slots {
		names[i] = s.Name
	}
	return names
}

func findSlot(slots []SlotDefinition, name string) *SlotDefinition {
	for _, s := range slots {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

func TestMerge_HooksFieldLevel(t *testing.T) {
	t.Run("disjoint hooks are both preserved", func(t *testing.T) {
		base := &Config{Hooks: HooksConfig{PreClear: []HookAction{{Type: "run", Command: "echo pre"}}}}
		override := &Config{Hooks: HooksConfig{PostMount: []HookAction{{Type: "run", Command: "echo post"}}}}

		got := Merge(base, override)

		assert.Equal(t, "echo pre", got.Hooks.PreClear[0].Command)
		assert.Equal(t, "echo post", got.Hooks.PostMount[0].Command)
	})

	t.Run("same hook field override wins", func(t *testing.T) {
		base := &Config{Hooks: HooksConfig{PostMount: []HookAction{{Type: "run", Command: "base-hook"}}}}
		override := &Config{Hooks: HooksConfig{PostMount: []HookAction{{Type: "run", Command: "override-hook"}}}}

		got := Merge(base, override)

		assert.Equal(t, "override-hook", got.Hooks.PostMount[0].Command)
	})
}

func boolPtr(b bool) *bool { return &b }

func TestMerge_LaunchShellOverride(t *testing.T) {
	t.Run("override true wins over base false", func(t *testing.T) {
		base := &Config{LaunchShell: boolPtr(false)}
		override := &Config{LaunchShell: boolPtr(true)}
		got := Merge(base, override)
		assert.Equal(t, boolPtr(true), got.LaunchShell)
	})

	t.Run("override false explicitly overrides base true", func(t *testing.T) {
		base := &Config{LaunchShell: boolPtr(true)}
		override := &Config{LaunchShell: boolPtr(false)}
		got := Merge(base, override)
		assert.Equal(t, boolPtr(false), got.LaunchShell)
	})

	t.Run("override nil does not override base true", func(t *testing.T) {
		base := &Config{LaunchShell: boolPtr(true)}
		override := &Config{LaunchShell: nil}
		got := Merge(base, override)
		assert.Equal(t, boolPtr(true), got.LaunchShell)
	})
}

func TestMerge_SlotEnvMerging(t *testing.T) {
	t.Run("slot env is preserved through merge", func(t *testing.T) {
		base := &Config{Slots: []SlotDefinition{
			{Name: "dev", Env: map[string]string{"PORT": "3001"}},
		}}
		override := &Config{Slots: nil}
		got := Merge(base, override)

		require.Len(t, got.Slots, 1)
		assert.Equal(t, map[string]string{"PORT": "3001"}, got.Slots[0].Env)
	})

	t.Run("override slot replaces env entirely", func(t *testing.T) {
		base := &Config{Slots: []SlotDefinition{
			{Name: "dev", Env: map[string]string{"PORT": "3001", "FOO": "bar"}},
		}}
		override := &Config{Slots: []SlotDefinition{
			{Name: "dev", Env: map[string]string{"PORT": "4000"}},
		}}
		got := Merge(base, override)

		require.Len(t, got.Slots, 1)
		assert.Equal(t, map[string]string{"PORT": "4000"}, got.Slots[0].Env)
	})
}

func TestMerge_BothEmpty(t *testing.T) {
	got := Merge(&Config{}, &Config{})

	assert.Empty(t, got.WtBasePath)
	assert.Nil(t, got.Slots)
}

func TestMerge_NilBase(t *testing.T) {
	override := &Config{
		WtBasePath: "/from-override",
		Slots:      slots("alpha"),
	}

	got := Merge(nil, override)

	require.NotNil(t, got)
	assert.Equal(t, "/from-override", got.WtBasePath)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, "alpha", got.Slots[0].Name)
}

func TestMerge_NilOverride(t *testing.T) {
	base := &Config{
		WtBasePath: "/from-base",
		Slots:      slots("beta"),
	}

	got := Merge(base, nil)

	require.NotNil(t, got)
	assert.Equal(t, "/from-base", got.WtBasePath)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, "beta", got.Slots[0].Name)
}

func TestMerge_BothNil(t *testing.T) {
	got := Merge(nil, nil)

	require.NotNil(t, got)
	assert.Empty(t, got.WtBasePath)
	assert.Nil(t, got.Slots)
}

func TestMerge_FullIntegration(t *testing.T) {
	base := &Config{
		WtBasePath: "/home/user/slots",
		Slots:      slots("dev", "staging"),
		Hooks: HooksConfig{
			PreMount:  []HookAction{{Type: "run", Command: "echo pre-mount"}},
			PostMount: []HookAction{{Type: "run", Command: "echo post-mount"}},
		},
	}
	override := &Config{
		WtBasePath: "/custom/slots",
		Slots:      []SlotDefinition{{Name: "dev", Icon: "🔥"}, {Name: "prod"}},
		Hooks: HooksConfig{
			PostMount: []HookAction{{Type: "run", Command: "custom-post-mount"}},
		},
	}

	got := Merge(base, override)

	assert.Equal(t, "/custom/slots", got.WtBasePath)
	require.Len(t, got.Slots, 3)
	assert.Equal(t, "🔥", findSlot(got.Slots, "dev").Icon)
	assert.NotNil(t, findSlot(got.Slots, "prod"))
	assert.Equal(t, "echo pre-mount", got.Hooks.PreMount[0].Command)
	assert.Equal(t, "custom-post-mount", got.Hooks.PostMount[0].Command)
}

func TestMerge_DoesNotMutateInputs(t *testing.T) {
	base := &Config{
		WtBasePath: "/base",
		Slots:      slots("a", "b"),
		Hooks:      HooksConfig{PreMount: []HookAction{{Type: "run", Command: "base-hook"}}},
	}
	override := &Config{
		WtBasePath: "/override",
		Slots:      slots("x"),
	}

	got := Merge(base, override)

	assert.Equal(t, "/base", base.WtBasePath)
	assert.Equal(t, "/override", override.WtBasePath)
	require.Len(t, base.Slots, 2)
	require.Len(t, override.Slots, 1)

	got.Slots[0].Name = "mutated"
	assert.Equal(t, "x", override.Slots[0].Name)
	assert.Equal(t, "a", base.Slots[0].Name)

}
