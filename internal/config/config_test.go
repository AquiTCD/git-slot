package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTOML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Config
	}{
		{
			name: "full valid TOML with all fields",
			input: `
wt_base_path = "/custom/path"

[[slots]]
name = "main-work"

[[slots]]
name = "hotfix"

[hooks]
pre_mount = [{type = "run", command = "scripts/pre.sh"}]
post_mount = [{type = "run", command = "scripts/post.sh"}]
pre_clear = [{type = "run", command = "scripts/pre-clear.sh"}]
post_clear = [{type = "run", command = "scripts/post-clear.sh"}]
`,
			expected: &Config{
				WtBasePath: "/custom/path",
				Slots: []SlotDefinition{
					{Name: "main-work"},
					{Name: "hotfix"},
				},
				Hooks: HooksConfig{
					PreMount:  []HookAction{{Type: "run", Command: "scripts/pre.sh"}},
					PostMount: []HookAction{{Type: "run", Command: "scripts/post.sh"}},
					PreClear:  []HookAction{{Type: "run", Command: "scripts/pre-clear.sh"}},
					PostClear: []HookAction{{Type: "run", Command: "scripts/post-clear.sh"}},
				},
			},
		},
		{
			name: "minimal TOML with only slots",
			input: `
[[slots]]
name = "work"
`,
			expected: &Config{
				Slots: []SlotDefinition{
					{Name: "work"},
				},
			},
		},
		{
			name: "wt_base_path set",
			input: `
wt_base_path = "/my/slots"

[[slots]]
name = "feature"
`,
			expected: &Config{
				WtBasePath: "/my/slots",
				Slots: []SlotDefinition{
					{Name: "feature"},
				},
			},
		},
		{
			name: "partial hooks with only post_mount",
			input: `
[[slots]]
name = "dev"

[hooks]
post_mount = [{type = "run", command = "run.sh"}]
`,
			expected: &Config{
				Slots: []SlotDefinition{
					{Name: "dev"},
				},
				Hooks: HooksConfig{
					PostMount: []HookAction{{Type: "run", Command: "run.sh"}},
				},
			},
		},
		{
			name: "slots with icon field",
			input: `
[[slots]]
name = "wood"
icon = "🌱"

[[slots]]
name = "fire"
icon = "🔥"

[[slots]]
name = "plain"
`,
			expected: &Config{
				Slots: []SlotDefinition{
					{Name: "wood", Icon: "🌱"},
					{Name: "fire", Icon: "🔥"},
					{Name: "plain"},
				},
			},
		},
		{
			name: "unknown keys silently ignored",
			input: `
unknown_key = "whatever"
wt_base_path = "~/wt"

[[slots]]
name = "s1"
extra = true
`,
			expected: &Config{
				WtBasePath: "~/wt",
				Slots: []SlotDefinition{
					{Name: "s1"},
				},
			},
		},
		{
			name:     "empty byte slice returns zero-value Config",
			input:    "",
			expected: &Config{},
		},
		{
			name: "slots with env field",
			input: `
[[slots]]
name = "main-work"

[slots.env]
PORT = "3001"
APP_ENV = "slot-dev"

[[slots]]
name = "hotfix"

[slots.env]
PORT = "3002"
`,
			expected: &Config{
				Slots: []SlotDefinition{
					{Name: "main-work", Env: map[string]string{"PORT": "3001", "APP_ENV": "slot-dev"}},
					{Name: "hotfix", Env: map[string]string{"PORT": "3002"}},
				},
			},
		},
		{
			name: "launch_shell setting",
			input: `
launch_shell = true

[[slots]]
name = "dev"
`,
			expected: &Config{
				LaunchShell: true,
				Slots: []SlotDefinition{
					{Name: "dev"},
				},
			},
		},
		{
			name: "slots without env field",
			input: `
[[slots]]
name = "plain"
`,
			expected: &Config{
				Slots: []SlotDefinition{
					{Name: "plain"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTOML([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseTOML_InvalidSyntax(t *testing.T) {
	input := []byte(`[invalid toml
this is = not valid`)

	_, err := ParseTOML(input)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigParse)
}

func TestConfig_FindSlot(t *testing.T) {
	cfg := &Config{
		Slots: []SlotDefinition{
			{Name: "dev", Icon: "🔧"},
			{Name: "staging"},
		},
	}

	t.Run("found", func(t *testing.T) {
		got := cfg.FindSlot("dev")
		require.NotNil(t, got)
		assert.Equal(t, "dev", got.Name)
		assert.Equal(t, "🔧", got.Icon)
	})

	t.Run("not found", func(t *testing.T) {
		got := cfg.FindSlot("nope")
		assert.Nil(t, got)
	})

	t.Run("empty config", func(t *testing.T) {
		empty := &Config{}
		got := empty.FindSlot("any")
		assert.Nil(t, got)
	})
}
