package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	toml "github.com/pelletier/go-toml/v2"
)

func resolveConfigPath(repoRoot string, global bool) (string, error) {
	if global {
		return config.DefaultGlobalConfigPath()
	}
	return filepath.Join(repoRoot, "git-slot.toml"), nil
}

func runHookHelper(a *app, out io.Writer, global bool) error {
	configPath, err := resolveConfigPath(a.repoRoot, global)
	if err != nil {
		return err
	}

	// Load existing config to pre-fill TUI
	var existingHooks []config.HookAction
	if data, err := os.ReadFile(configPath); err == nil {
		if cfg, err := config.ParseTOML(data); err == nil {
			existingHooks = cfg.Hooks.PostMount
		}
	}

	wt := git.NewExecWorktree(a.repoRoot)

	// List ignored files
	files, err := wt.ListIgnoredFiles()
	if err != nil {
		return fmt.Errorf("failed to list ignored files: %w", err)
	}

	var items []tui.HookItem
	for _, f := range files {
		action := tui.ActionNone
		// Check if this file already has a hook configured
		potentialSrc := filepath.Join("$GSL_REPO_ROOT", f.Path)
		for _, h := range existingHooks {
			if h.Source == potentialSrc {
				switch h.Type {
				case "link":
					action = tui.ActionSymlink
				case "copy":
					action = tui.ActionCopy
				}
				break
			}
		}
		items = append(items, tui.HookItem{Path: f.Path, Action: action})
	}

	if len(items) == 0 {
		_, _ = fmt.Fprintln(out, "No ignored files found to setup hooks for.")
		return nil
	}

	noColor := tui.IsNoColor()
	model := tui.NewHookModelFromItems(items, noColor)

	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("hook helper tui: %w", err)
	}

	m := finalModel.(tui.HookModel)
	if m.Aborted() {
		return nil
	}

	results := m.GetResults()
	return updateConfigWithHooks(a, results, out, global)
}

const hookManagedMarker = "# --- Managed by git-slot --hook ---"

func buildPostMountHooks(existingHooks []config.HookAction, items []tui.HookItem) ([]config.HookAction, bool) {
	var hooks []config.HookAction
	for _, existing := range existingHooks {
		if existing.Type == "run" {
			hooks = append(hooks, existing)
		}
	}

	hasNew := false
	for _, item := range items {
		if item.Action == tui.ActionNone {
			continue
		}
		hasNew = true
		action := config.HookAction{
			Source: filepath.Join("$GSL_REPO_ROOT", item.Path),
			Dest:   filepath.Join("$GSL_SLOT_PATH", item.Path),
		}
		switch item.Action {
		case tui.ActionSymlink:
			action.Type = "link"
		case tui.ActionCopy:
			action.Type = "copy"
		}
		hooks = append(hooks, action)
	}

	return hooks, hasNew
}

func generateHooksTOML(originalContent string, hooks []config.HookAction) (string, error) {
	if idx := strings.Index(originalContent, hookManagedMarker); idx != -1 {
		originalContent = originalContent[:idx]
	}

	re := regexp.MustCompile(`(?ms)^\s*\[\[hooks\.post_mount\]\].*?(\n\n|(?:\n\s*\[)|\z)`)
	contentWithoutHooks := re.ReplaceAllString(originalContent, "$1")
	contentWithoutHooks = strings.TrimSpace(contentWithoutHooks)

	type postMountWrapper struct {
		PostMount []config.HookAction `toml:"post_mount"`
	}
	type hooksWrapper struct {
		Hooks postMountWrapper `toml:"hooks"`
	}

	newHooksData, err := toml.Marshal(hooksWrapper{
		Hooks: postMountWrapper{PostMount: hooks},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal hooks: %w", err)
	}

	var out strings.Builder
	out.WriteString(contentWithoutHooks)
	out.WriteString("\n\n")
	out.WriteString(hookManagedMarker + "\n")
	out.WriteString(string(newHooksData))

	return out.String(), nil
}

func updateConfigWithHooks(a *app, items []tui.HookItem, out io.Writer, global bool) error {
	configPath, err := resolveConfigPath(a.repoRoot, global)
	if err != nil {
		return err
	}

	var originalContent string
	if data, err := os.ReadFile(configPath); err == nil {
		originalContent = string(data)
	}

	var existingHooks []config.HookAction
	if originalContent != "" {
		if cfg, err := config.ParseTOML([]byte(originalContent)); err == nil {
			existingHooks = cfg.Hooks.PostMount
		}
	}

	hooks, hasNew := buildPostMountHooks(existingHooks, items)
	if !hasNew && len(hooks) == 0 {
		_, _ = fmt.Fprintln(out, "No actions selected and no existing hooks found.")
		return nil
	}

	finalContent, err := generateHooksTOML(originalContent, hooks)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(configPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	loc := "project"
	if global {
		loc = "global"
	}
	_, _ = fmt.Fprintf(out, "Successfully updated %s git-slot.toml while preserving comments! 💎\n", loc)
	return nil
}
