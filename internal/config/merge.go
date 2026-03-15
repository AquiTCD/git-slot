package config

func Merge(base, override *Config) *Config {
	if base == nil && override == nil {
		return &Config{}
	}
	if base == nil {
		return copyConfig(override)
	}
	if override == nil {
		return copyConfig(base)
	}

	merged := &Config{
		GwqBaseDir:    base.GwqBaseDir,
		SlotsBasePath: base.SlotsBasePath,
		Slots:         copySlots(base.Slots),
		Hooks:         base.Hooks,
		TUI:           base.TUI,
	}

	if override.GwqBaseDir != "" {
		merged.GwqBaseDir = override.GwqBaseDir
	}
	if override.SlotsBasePath != "" {
		merged.SlotsBasePath = override.SlotsBasePath
	}

	if override.Slots != nil {
		merged.Slots = copySlots(override.Slots)
	}

	if override.Hooks.PreLoad != "" {
		merged.Hooks.PreLoad = override.Hooks.PreLoad
	}
	if override.Hooks.PostLoad != "" {
		merged.Hooks.PostLoad = override.Hooks.PostLoad
	}
	if override.Hooks.PreClear != "" {
		merged.Hooks.PreClear = override.Hooks.PreClear
	}
	if override.Hooks.PostClear != "" {
		merged.Hooks.PostClear = override.Hooks.PostClear
	}

	if override.TUI.Filter {
		merged.TUI.Filter = override.TUI.Filter
	}

	return merged
}

func copyConfig(src *Config) *Config {
	dst := &Config{
		GwqBaseDir:    src.GwqBaseDir,
		SlotsBasePath: src.SlotsBasePath,
		Slots:         copySlots(src.Slots),
		Hooks:         src.Hooks,
		TUI:           src.TUI,
	}
	return dst
}

func copySlots(src []SlotDefinition) []SlotDefinition {
	if src == nil {
		return nil
	}
	dst := make([]SlotDefinition, len(src))
	copy(dst, src)
	return dst
}
