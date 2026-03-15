package config

import (
	"errors"
	"fmt"

	"github.com/AquiTCD/git-slot/internal/errutil"
	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	GwqBaseDir    string           `toml:"gwq_basedir"`
	SlotsBasePath string           `toml:"slots_base_path"`
	Slots         []SlotDefinition `toml:"slots"`
	Hooks         HooksConfig      `toml:"hooks"`
}

type SlotDefinition struct {
	Name string `toml:"name"`
	Icon string `toml:"icon"`
}

type HooksConfig struct {
	PreLoad   string `toml:"pre_load"`
	PostLoad  string `toml:"post_load"`
	PreClear  string `toml:"pre_clear"`
	PostClear string `toml:"post_clear"`
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
