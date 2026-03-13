package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func slots(names ...string) []SlotDefinition {
	s := make([]SlotDefinition, len(names))
	for i, n := range names {
		s[i] = SlotDefinition{Name: n}
	}
	return s
}

func TestValidate_ValidConfigs(t *testing.T) {
	tests := []struct {
		name  string
		slots []SlotDefinition
	}{
		{"three unique slots", slots("dev", "staging", "prod")},
		{"single slot", slots("only")},
		{"alphanumeric with hyphen", slots("slot-1")},
		{"alphanumeric with underscore", slots("foo_bar")},
		{"single char lowercase", slots("a")},
		{"mixed case with hyphen", slots("A-z")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Slots: tt.slots}
			err := Validate(cfg)
			assert.NoError(t, err)
		})
	}
}

func TestValidate_EmptySlots(t *testing.T) {
	cfg := &Config{Slots: []SlotDefinition{}}
	err := Validate(cfg)
	require.ErrorIs(t, err, ErrNoSlots)
}

func TestValidate_NilSlots(t *testing.T) {
	cfg := &Config{Slots: nil}
	err := Validate(cfg)
	require.ErrorIs(t, err, ErrNoSlots)
}

func TestValidate_EmptySlotName(t *testing.T) {
	cfg := &Config{Slots: slots("good", "")}
	err := Validate(cfg)
	require.ErrorIs(t, err, ErrEmptySlotName)
}

func TestValidate_WhitespaceOnlySlotName(t *testing.T) {
	cfg := &Config{Slots: slots("good", "   ")}
	err := Validate(cfg)
	require.ErrorIs(t, err, ErrEmptySlotName)
}

func TestValidate_DuplicateSlotNames(t *testing.T) {
	cfg := &Config{Slots: slots("dev", "staging", "dev")}
	err := Validate(cfg)

	var target *ErrDuplicateSlotName
	require.ErrorAs(t, err, &target)
	assert.Equal(t, "dev", target.Name)
}

func TestValidate_InvalidCharacters(t *testing.T) {
	cfg := &Config{Slots: slots("my slot!")}
	err := Validate(cfg)

	var target *ErrInvalidSlotName
	require.ErrorAs(t, err, &target)
	assert.Equal(t, "my slot!", target.Name)
}

func TestValidate_ReservedName(t *testing.T) {
	cfg := &Config{Slots: slots(".swap-temp")}
	err := Validate(cfg)

	var target *ErrReservedSlotName
	require.ErrorAs(t, err, &target)
	assert.Equal(t, ".swap-temp", target.Name)
}

func TestValidate_UnicodeName(t *testing.T) {
	cfg := &Config{Slots: slots("スロット")}
	err := Validate(cfg)

	var target *ErrInvalidSlotName
	require.ErrorAs(t, err, &target)
	assert.Equal(t, "スロット", target.Name)
}
