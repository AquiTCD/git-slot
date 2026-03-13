package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempConfig(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

const globalTOML = `
gwq_basedir = "~/worktrees"

[[slots]]
name = "dev"

[[slots]]
name = "staging"

[hooks]
pre_load = "global-pre.sh"
post_load = "global-post.sh"
`

const projectTOML = `
gwq_basedir = "~/project-trees"

[[slots]]
name = "feature"

[hooks]
post_load = "project-post.sh"
pre_clear = "project-pre-clear.sh"
`

func TestLoadConfig_BothExist(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)
	pp := writeTempConfig(t, dir, "project.toml", projectTOML)

	cfg, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.NoError(t, err)

	assert.Equal(t, "~/project-trees", cfg.GwqBaseDir)
	assert.Equal(t, []SlotDefinition{{Name: "feature"}}, cfg.Slots)
	assert.Equal(t, "global-pre.sh", cfg.Hooks.PreLoad)
	assert.Equal(t, "project-post.sh", cfg.Hooks.PostLoad)
	assert.Equal(t, "project-pre-clear.sh", cfg.Hooks.PreClear)
}

func TestLoadConfig_OnlyGlobal(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  gp,
		ProjectPath: filepath.Join(dir, "nonexistent.toml"),
	})
	require.NoError(t, err)

	assert.Equal(t, "~/worktrees", cfg.GwqBaseDir)
	assert.Equal(t, []SlotDefinition{{Name: "dev"}, {Name: "staging"}}, cfg.Slots)
}

func TestLoadConfig_OnlyProject(t *testing.T) {
	dir := t.TempDir()
	pp := writeTempConfig(t, dir, "project.toml", projectTOML)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  filepath.Join(dir, "nonexistent.toml"),
		ProjectPath: pp,
	})
	require.NoError(t, err)

	assert.Equal(t, "~/project-trees", cfg.GwqBaseDir)
	assert.Equal(t, []SlotDefinition{{Name: "feature"}}, cfg.Slots)
}

func TestLoadConfig_NeitherExists(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadConfig(LoadOptions{
		GlobalPath:  filepath.Join(dir, "nope.toml"),
		ProjectPath: filepath.Join(dir, "also-nope.toml"),
	})
	require.ErrorIs(t, err, ErrNoConfig)
}

func TestLoadConfig_ProjectTOMLSyntaxError(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)
	pp := writeTempConfig(t, dir, "bad-project.toml", `[invalid toml
this is = broken`)

	_, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigParse)
	assert.Contains(t, err.Error(), "bad-project.toml")
}

func TestLoadConfig_GlobalTOMLSyntaxError(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "bad-global.toml", `[invalid toml
this is = broken`)
	pp := writeTempConfig(t, dir, "project.toml", projectTOML)

	_, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigParse)
	assert.Contains(t, err.Error(), "bad-global.toml")
}

func TestLoadConfig_MergedValidationError(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
[[slots]]
name = "dup"
`)
	pp := writeTempConfig(t, dir, "project.toml", `
[[slots]]
name = "dup"

[[slots]]
name = "dup"
`)

	_, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.Error(t, err)

	var target *ErrDuplicateSlotName
	assert.ErrorAs(t, err, &target)
}

func TestLoadConfig_ProjectOverridesGwqBaseDir(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
gwq_basedir = "~/global-base"

[[slots]]
name = "s1"
`)
	pp := writeTempConfig(t, dir, "project.toml", `
gwq_basedir = "~/project-base"

[[slots]]
name = "s1"
`)

	cfg, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.NoError(t, err)
	assert.Equal(t, "~/project-base", cfg.GwqBaseDir)
}

func TestLoadConfig_ProjectSlotsReplaceGlobal(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
[[slots]]
name = "global-slot"
`)
	pp := writeTempConfig(t, dir, "project.toml", `
[[slots]]
name = "project-slot"
`)

	cfg, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.NoError(t, err)
	assert.Equal(t, []SlotDefinition{{Name: "project-slot"}}, cfg.Slots)
}

func TestLoadConfig_HooksFieldMerged(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
[[slots]]
name = "s1"

[hooks]
pre_load = "global-pre.sh"
post_load = "global-post.sh"
`)
	pp := writeTempConfig(t, dir, "project.toml", `
[[slots]]
name = "s1"

[hooks]
post_load = "project-post.sh"
pre_clear = "project-pre-clear.sh"
`)

	cfg, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.NoError(t, err)
	assert.Equal(t, "global-pre.sh", cfg.Hooks.PreLoad)
	assert.Equal(t, "project-post.sh", cfg.Hooks.PostLoad)
	assert.Equal(t, "project-pre-clear.sh", cfg.Hooks.PreClear)
	assert.Empty(t, cfg.Hooks.PostClear)
}

func TestLoadConfig_EmptyFileFailsValidation(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "empty.toml", "")

	_, err := LoadConfig(LoadOptions{
		GlobalPath:  gp,
		ProjectPath: filepath.Join(dir, "nonexistent.toml"),
	})
	require.ErrorIs(t, err, ErrNoSlots)
}

func TestLoadConfig_RepoRootDeriveProjectPath(t *testing.T) {
	dir := t.TempDir()
	repoRoot := filepath.Join(dir, "repo")
	require.NoError(t, os.MkdirAll(repoRoot, 0755))

	writeTempConfig(t, repoRoot, "git-slot.toml", `
[[slots]]
name = "auto-discovered"
`)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath: filepath.Join(dir, "nonexistent.toml"),
		RepoRoot:   repoRoot,
	})
	require.NoError(t, err)
	assert.Equal(t, []SlotDefinition{{Name: "auto-discovered"}}, cfg.Slots)
}

func TestDefaultGlobalConfigPath(t *testing.T) {
	p, err := DefaultGlobalConfigPath()
	require.NoError(t, err)
	assert.Contains(t, p, filepath.Join(".config", "git-slot", "config.toml"))
}
