package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AquiTCD/git-slot/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestAggregateItems_BelowThreshold(t *testing.T) {
	items := []tui.HookItem{
		{Path: ".env"},
		{Path: "vendor/"},
		{Path: "node_modules/"},
	}
	result := aggregateItems(items, 3)
	require.Len(t, result, 3)
	assert.Equal(t, ".env", result[0].Path)
}

func TestAggregateItems_AboveThreshold(t *testing.T) {
	items := []tui.HookItem{
		{Path: "storage/a/"},
		{Path: "storage/b/"},
		{Path: "storage/c/"},
		{Path: "storage/d/"},
		{Path: ".env"},
	}
	result := aggregateItems(items, 3)
	require.Len(t, result, 2)

	assert.Equal(t, "storage/", result[0].Path)
	assert.True(t, result[0].IsDir)
	assert.Equal(t, 4, result[0].ChildCount)
	require.Len(t, result[0].Children, 4)
	assert.Equal(t, "storage/a/", result[0].Children[0].Path)

	assert.Equal(t, ".env", result[1].Path)
}

func TestAggregateItems_PreservesOriginalDir(t *testing.T) {
	items := []tui.HookItem{
		{Path: "vendor/", IsDir: true},
		{Path: ".env"},
	}
	result := aggregateItems(items, 3)
	require.Len(t, result, 2)
	assert.Equal(t, "vendor/", result[0].Path)
	assert.True(t, result[0].IsDir)
	assert.Equal(t, 0, result[0].ChildCount)
}

func TestAggregateItems_NestedAggregation(t *testing.T) {
	items := []tui.HookItem{
		{Path: "cache/a/x"},
		{Path: "cache/a/y"},
		{Path: "cache/a/z"},
		{Path: "cache/b/x"},
	}
	result := aggregateItems(items, 3)
	require.Len(t, result, 1)
	assert.Equal(t, "cache/", result[0].Path)
	assert.Equal(t, 4, result[0].ChildCount)
}

func TestAggregateItems_MixedDepths(t *testing.T) {
	items := []tui.HookItem{
		{Path: "dist/foo.dmg"},
		{Path: "dist/bar.zip"},
		{Path: "dist/baz.tar"},
		{Path: "dist/sub/deep.txt"},
		{Path: ".env"},
		{Path: "tmp/"},
	}
	result := aggregateItems(items, 3)
	require.Len(t, result, 3)
	assert.Equal(t, "dist/", result[0].Path)
	assert.Equal(t, 4, result[0].ChildCount)
	assert.Equal(t, ".env", result[1].Path)
	assert.Equal(t, "tmp/", result[2].Path)
}
