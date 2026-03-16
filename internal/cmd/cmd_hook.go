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

func runHookHelper(a *app, out io.Writer, global bool) error {
	var configPath string
	if global {
		p, err := config.DefaultGlobalConfigPath()
		if err != nil {
			return err
		}
		configPath = p
	} else {
		configPath = filepath.Join(a.repoRoot, "git-slot.toml")
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

func updateConfigWithHooks(a *app, items []tui.HookItem, out io.Writer, global bool) error {
	var configPath string
	if global {
		p, err := config.DefaultGlobalConfigPath()
		if err != nil {
			return err
		}
		configPath = p
	} else {
		configPath = filepath.Join(a.repoRoot, "git-slot.toml")
	}

	// Read existing content to preserve comments
	var originalContent string
	if data, err := os.ReadFile(configPath); err == nil {
		originalContent = string(data)
	}

	// Parse to find existing 'run' hooks we might want to preserve
	// (Simple approach: we'll re-generate the entire post_mount section,
	// but we could be more surgical if needed).
	var targetCfg *config.Config
	if originalContent != "" {
		if cfg, err := config.ParseTOML([]byte(originalContent)); err == nil {
			targetCfg = cfg
		}
	}
	if targetCfg == nil {
		targetCfg = &config.Config{}
	}

	// Filter out automatic link/copy actions and keep manual ones
	var finalPostMount []config.HookAction
	for _, existing := range targetCfg.Hooks.PostMount {
		if existing.Type == "run" {
			finalPostMount = append(finalPostMount, existing)
		}
	}

	// Add new items from TUI
	found := false
	for _, item := range items {
		if item.Action == tui.ActionNone {
			continue
		}
		found = true
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
		finalPostMount = append(finalPostMount, action)
	}

	if !found && len(finalPostMount) == 0 {
		_, _ = fmt.Fprintln(out, "No actions selected and no existing hooks found.")
		return nil
	}

	// SURGERY: First, remove the entire managed block if it exists
	marker := "# --- Managed by git-slot --hook ---"
	if idx := strings.Index(originalContent, marker); idx != -1 {
		originalContent = originalContent[:idx]
	}

	// Then, remove any remaining [[hooks.post_mount]] blocks (for robustness)
	re := regexp.MustCompile(`(?ms)^\s*\[\[hooks\.post_mount\]\].*?(\n\n|(?:\n\s*\[)|\z)`)
	contentWithoutHooks := re.ReplaceAllString(originalContent, "$1")
	contentWithoutHooks = strings.TrimSpace(contentWithoutHooks)

	// Clean up potential trailing empty [hooks] if it's now empty (optional, keeping it simple for now)

	// Generate the new hooks TOML
	// We want to generate [[hooks.post_mount]] blocks directly.
	// Since go-toml/v2 doesn't easily support un-rooted slices with custom names,
	// we'll use a temporary map or similar to get the right format.
	type postMountWrapper struct {
		PostMount []config.HookAction `toml:"post_mount"`
	}
	type hooksWrapper struct {
		Hooks postMountWrapper `toml:"hooks"`
	}
	
	wrapper := hooksWrapper{
		Hooks: postMountWrapper{
			PostMount: finalPostMount,
		},
	}
	
	newHooksData, err := toml.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal hooks: %w", err)
	}

	// Reassemble: 1. Processed original (with comments!) + 2. New Hooks
	var finalOutput strings.Builder
	finalOutput.WriteString(contentWithoutHooks)
	finalOutput.WriteString("\n\n")
	finalOutput.WriteString("# --- Managed by git-slot --hook ---\n")
	finalOutput.WriteString(string(newHooksData))

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(configPath, []byte(finalOutput.String()), 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	loc := "project"
	if global {
		loc = "global"
	}
	_, _ = fmt.Fprintf(out, "Successfully updated %s git-slot.toml while preserving comments! 💎\n", loc)
	return nil
}
