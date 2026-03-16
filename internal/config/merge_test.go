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
			name:     "override slots_base_path wins",
			base:     &Config{SlotsBasePath: "/base/slots"},
			override: &Config{SlotsBasePath: "/override/slots"},
			wantPath: "/override/slots",
		},
		{
			name:     "empty override preserves base slots_base_path",
			base:     &Config{SlotsBasePath: "/base/slots"},
			override: &Config{},
			wantPath: "/base/slots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.override)
			if tt.wantPath != "" {
				assert.Equal(t, tt.wantPath, got.SlotsBasePath)
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

func TestMerge_BothEmpty(t *testing.T) {
	got := Merge(&Config{}, &Config{})

	assert.Empty(t, got.SlotsBasePath)
	assert.Nil(t, got.Slots)
}

func TestMerge_NilBase(t *testing.T) {
	override := &Config{
		SlotsBasePath: "/from-override",
		Slots:         slots("alpha"),
	}

	got := Merge(nil, override)

	require.NotNil(t, got)
	assert.Equal(t, "/from-override", got.SlotsBasePath)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, "alpha", got.Slots[0].Name)
}

func TestMerge_NilOverride(t *testing.T) {
	base := &Config{
		SlotsBasePath: "/from-base",
		Slots:         slots("beta"),
	}

	got := Merge(base, nil)

	require.NotNil(t, got)
	assert.Equal(t, "/from-base", got.SlotsBasePath)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, "beta", got.Slots[0].Name)
}

func TestMerge_BothNil(t *testing.T) {
	got := Merge(nil, nil)

	require.NotNil(t, got)
	assert.Empty(t, got.SlotsBasePath)
	assert.Nil(t, got.Slots)
}

func TestMerge_FullIntegration(t *testing.T) {
	base := &Config{
		SlotsBasePath: "/home/user/slots",
		Slots:         slots("dev", "staging"),
		Hooks: HooksConfig{
			PreMount:  []HookAction{{Type: "run", Command: "echo pre-mount"}},
			PostMount: []HookAction{{Type: "run", Command: "echo post-mount"}},
		},
	}
	override := &Config{
		SlotsBasePath: "/custom/slots",
		Slots:         []SlotDefinition{{Name: "dev", Icon: "🔥"}, {Name: "prod"}},
		Hooks: HooksConfig{
			PostMount: []HookAction{{Type: "run", Command: "custom-post-mount"}},
		},
	}

	got := Merge(base, override)

	assert.Equal(t, "/custom/slots", got.SlotsBasePath)
	require.Len(t, got.Slots, 3)
	assert.Equal(t, "🔥", findSlot(got.Slots, "dev").Icon)
	assert.NotNil(t, findSlot(got.Slots, "prod"))
	assert.Equal(t, "echo pre-mount", got.Hooks.PreMount[0].Command)
	assert.Equal(t, "custom-post-mount", got.Hooks.PostMount[0].Command)
}

func TestMerge_DoesNotMutateInputs(t *testing.T) {
	base := &Config{
		SlotsBasePath: "/base",
		Slots:         slots("a", "b"),
		Hooks:         HooksConfig{PreMount: []HookAction{{Type: "run", Command: "base-hook"}}},
	}
	override := &Config{
		SlotsBasePath: "/override",
		Slots:         slots("x"),
	}

	got := Merge(base, override)

	assert.Equal(t, "/base", base.SlotsBasePath)
	assert.Equal(t, "/override", override.SlotsBasePath)
	require.Len(t, base.Slots, 2)
	require.Len(t, override.Slots, 1)

	got.Slots[0].Name = "mutated"
	assert.Equal(t, "x", override.Slots[0].Name)
	assert.Equal(t, "a", base.Slots[0].Name)

}
