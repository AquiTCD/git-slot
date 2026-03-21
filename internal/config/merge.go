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
		WtBasePath:  base.WtBasePath,
		LaunchShell: copyBoolPtr(base.LaunchShell),
		Slots:       copySlots(base.Slots),
		Hooks:       copyHooks(base.Hooks),
	}

	if override.WtBasePath != "" {
		merged.WtBasePath = override.WtBasePath
	}

	if override.LaunchShell != nil {
		merged.LaunchShell = copyBoolPtr(override.LaunchShell)
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

	if override.Hooks.PreMount != nil {
		merged.Hooks.PreMount = copyHookActions(override.Hooks.PreMount)
	}
	if override.Hooks.PostMount != nil {
		merged.Hooks.PostMount = copyHookActions(override.Hooks.PostMount)
	}
	if override.Hooks.PreClear != nil {
		merged.Hooks.PreClear = copyHookActions(override.Hooks.PreClear)
	}
	if override.Hooks.PostClear != nil {
		merged.Hooks.PostClear = copyHookActions(override.Hooks.PostClear)
	}

	return merged
}

func copyConfig(src *Config) *Config {
	dst := &Config{
		WtBasePath:  src.WtBasePath,
		LaunchShell: copyBoolPtr(src.LaunchShell),
		Slots:       copySlots(src.Slots),
		Hooks:       copyHooks(src.Hooks),
	}
	return dst
}

func copyBoolPtr(b *bool) *bool {
	if b == nil {
		return nil
	}
	v := *b
	return &v
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
	for i, s := range src {
		dst[i] = SlotDefinition{
			Name: s.Name,
			Icon: s.Icon,
			Env:  copyEnvMap(s.Env),
		}
	}
	return dst
}

func copyEnvMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
