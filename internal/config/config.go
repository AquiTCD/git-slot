package config

import (
	"errors"
	"fmt"

	"github.com/AquiTCD/git-slot/internal/errutil"
	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	SlotsBasePath string           `toml:"slots_base_path,omitempty"`
	LaunchShell   bool             `toml:"launch_shell,omitempty"`
	Slots         []SlotDefinition `toml:"slots,omitempty"`
	Hooks         HooksConfig      `toml:"hooks,omitempty"`
}

type SlotDefinition struct {
	Name string            `toml:"name,omitempty"`
	Icon string            `toml:"icon,omitempty"`
	Env  map[string]string `toml:"env,omitempty"`
}

type HookAction struct {
	Type    string `toml:"type,omitempty"`    // "link", "copy", "run"
	Source  string `toml:"source,omitempty"`  // For link/copy
	Dest    string `toml:"dest,omitempty"`    // For link/copy
	Command string `toml:"command,omitempty"` // For run
}

type HooksConfig struct {
	PreMount  []HookAction `toml:"pre_mount,omitempty"`
	PostMount []HookAction `toml:"post_mount,omitempty"`
	PreClear  []HookAction `toml:"pre_clear,omitempty"`
	PostClear []HookAction `toml:"post_clear,omitempty"`
}

var (
	ErrNoSlots       = errors.New("no slots defined")
	ErrEmptySlotName = errors.New("slot name is empty")
	ErrConfigParse   = errutil.NewExitError("failed to parse configuration", 2)
)

func ParseTOML(data []byte) (*Config, error) {
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", err.Error(), ErrConfigParse)
	}
	return &cfg, nil
}

type ErrDuplicateSlotName struct {
	Name string
}

func (e *ErrDuplicateSlotName) Error() string {
	return fmt.Sprintf("duplicate slot name: %q", e.Name)
}

type ErrInvalidSlotName struct {
	Name string
}

func (e *ErrInvalidSlotName) Error() string {
	return fmt.Sprintf("invalid slot name: %q", e.Name)
}

type ErrReservedSlotName struct {
	Name string
}

func (e *ErrReservedSlotName) Error() string {
	return fmt.Sprintf("reserved slot name: %q", e.Name)
}
