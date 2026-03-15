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
gwq_basedir = "~/worktrees"
slots_base_path = "/custom/path"

[[slots]]
name = "main-work"

[[slots]]
name = "hotfix"

[hooks]
pre_load = "scripts/pre.sh"
post_load = "scripts/post.sh"
pre_clear = "scripts/pre-clear.sh"
post_clear = "scripts/post-clear.sh"
`,
			expected: &Config{
				GwqBaseDir:    "~/worktrees",
				SlotsBasePath: "/custom/path",
				Slots: []SlotDefinition{
					{Name: "main-work"},
					{Name: "hotfix"},
				},
				Hooks: HooksConfig{
					PreLoad:   "scripts/pre.sh",
					PostLoad:  "scripts/post.sh",
					PreClear:  "scripts/pre-clear.sh",
					PostClear: "scripts/post-clear.sh",
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
			name: "gwq_basedir and slots only",
			input: `
gwq_basedir = "~/trees"

[[slots]]
name = "dev"
`,
			expected: &Config{
				GwqBaseDir: "~/trees",
				Slots: []SlotDefinition{
					{Name: "dev"},
				},
			},
		},
		{
			name: "slots_base_path set",
			input: `
slots_base_path = "/my/slots"

[[slots]]
name = "feature"
`,
			expected: &Config{
				SlotsBasePath: "/my/slots",
				Slots: []SlotDefinition{
					{Name: "feature"},
				},
			},
		},
		{
			name: "partial hooks with only post_load",
			input: `
[[slots]]
name = "dev"

[hooks]
post_load = "run.sh"
`,
			expected: &Config{
				Slots: []SlotDefinition{
					{Name: "dev"},
				},
				Hooks: HooksConfig{
					PostLoad: "run.sh",
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
gwq_basedir = "~/wt"

[[slots]]
name = "s1"
extra = true
`,
			expected: &Config{
				GwqBaseDir: "~/wt",
				Slots: []SlotDefinition{
					{Name: "s1"},
				},
			},
		},
		{
			name: "tui filter enabled",
			input: `
[[slots]]
name = "dev"

[tui]
filter = true
`,
			expected: &Config{
				Slots: []SlotDefinition{
					{Name: "dev"},
				},
				TUI: TUIConfig{Filter: true},
			},
		},
		{
			name:     "empty byte slice returns zero-value Config",
			input:    "",
			expected: &Config{},
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
