package config

import (
	"regexp"
	"strings"
)

var validSlotName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var reservedNames = map[string]bool{}

func Validate(cfg *Config) error {
	if len(cfg.Slots) == 0 {
		return ErrNoSlots
	}

	seen := make(map[string]bool, len(cfg.Slots))

	for i := range cfg.Slots {
		name := strings.TrimSpace(cfg.Slots[i].Name)
		cfg.Slots[i].Name = name

		if name == "" {
			return ErrEmptySlotName
		}

		if reservedNames[name] {
			return &ErrReservedSlotName{Name: name}
		}

		if !validSlotName.MatchString(name) {
			return &ErrInvalidSlotName{Name: name}
		}

		if seen[name] {
			return &ErrDuplicateSlotName{Name: name}
		}
		seen[name] = true
	}

	return nil
}
