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
		SlotsBasePath: base.SlotsBasePath,
		Slots:         copySlots(base.Slots),
		Hooks:         copyHooks(base.Hooks),
	}

	if override.SlotsBasePath != "" {
		merged.SlotsBasePath = override.SlotsBasePath
	}

	for _, overrideSlot := range override.Slots {
		found := false
		for i, baseSlot := range merged.Slots {
			if baseSlot.Name == overrideSlot.Name {
				merged.Slots[i] = overrideSlot
				found = true
				break
			}
		}
		if !found {
			merged.Slots = append(merged.Slots, overrideSlot)
		}
	}

	if len(override.Hooks.PreMount) > 0 {
		merged.Hooks.PreMount = copyHookActions(override.Hooks.PreMount)
	}
	if len(override.Hooks.PostMount) > 0 {
		merged.Hooks.PostMount = copyHookActions(override.Hooks.PostMount)
	}
	if len(override.Hooks.PreClear) > 0 {
		merged.Hooks.PreClear = copyHookActions(override.Hooks.PreClear)
	}
	if len(override.Hooks.PostClear) > 0 {
		merged.Hooks.PostClear = copyHookActions(override.Hooks.PostClear)
	}

	return merged
}

func copyConfig(src *Config) *Config {
	dst := &Config{
		SlotsBasePath: src.SlotsBasePath,
		Slots:         copySlots(src.Slots),
		Hooks:         copyHooks(src.Hooks),
	}
	return dst
}

func copyHooks(src HooksConfig) HooksConfig {
	return HooksConfig{
		PreMount:  copyHookActions(src.PreMount),
		PostMount: copyHookActions(src.PostMount),
		PreClear:  copyHookActions(src.PreClear),
		PostClear: copyHookActions(src.PostClear),
	}
}

func copyHookActions(src []HookAction) []HookAction {
	if src == nil {
		return nil
	}
	dst := make([]HookAction, len(src))
	copy(dst, src)
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
