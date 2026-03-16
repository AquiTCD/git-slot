package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AquiTCD/git-slot/internal/tui"
)

func TestUpdateConfigWithHooks(t *testing.T) {
	repoRoot := t.TempDir()
	configPath := filepath.Join(repoRoot, "git-slot.toml")

	items := []tui.HookItem{
		{Path: ".env", Action: tui.ActionSymlink},
		{Path: "config/local.json", Action: tui.ActionCopy},
		{Path: "ignored.txt", Action: tui.ActionNone},
	}

	a := &app{repoRoot: repoRoot}
	var out strings.Builder
	err := updateConfigWithHooks(a, items, &out, false)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	tomlContent := string(content)

	// Check content
	expectedFragments := []string{
		"type = 'link'",
		"source = '$GSL_REPO_ROOT/.env'",
		"dest = '$GSL_SLOT_PATH/.env'",
		"type = 'copy'",
		"source = '$GSL_REPO_ROOT/config/local.json'",
		"dest = '$GSL_SLOT_PATH/config/local.json'",
	}

	for _, frag := range expectedFragments {
		if !strings.Contains(tomlContent, frag) {
			t.Errorf("Expected fragment missing from TOML: %s", frag)
		}
	}

	if strings.Contains(tomlContent, "ignored.txt") {
		t.Errorf("ActionNone item should not be in the TOML")
	}
}
