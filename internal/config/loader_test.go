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
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
	return path
}

const globalTOML = `
wt_base_path = "~/worktrees"

[[slots]]
name = "dev"

[[slots]]
name = "staging"

[hooks]
pre_mount = [{type = "run", command = "global-pre.sh"}]
post_mount = [{type = "run", command = "global-post.sh"}]
`

const projectTOML = `
wt_base_path = "~/project-trees"

[[slots]]
name = "feature"

[hooks]
post_mount = [{type = "run", command = "project-post.sh"}]
pre_clear = [{type = "run", command = "project-pre-clear.sh"}]
`

func TestLoadConfig_BothExist(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)
	pp := writeTempConfig(t, dir, "project.toml", projectTOML)

	cfg, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.NoError(t, err)

	assert.Equal(t, "~/project-trees", cfg.WtBasePath)
	// Now slots are merged/appended
	assert.ElementsMatch(t, []SlotDefinition{{Name: "dev"}, {Name: "staging"}, {Name: "feature"}}, cfg.Slots)
	assert.Equal(t, []HookAction{{Type: "run", Command: "global-pre.sh"}}, cfg.Hooks.PreMount)
	assert.Equal(t, []HookAction{{Type: "run", Command: "project-post.sh"}}, cfg.Hooks.PostMount)
	assert.Equal(t, []HookAction{{Type: "run", Command: "project-pre-clear.sh"}}, cfg.Hooks.PreClear)
}

func TestLoadConfig_OnlyGlobal(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  gp,
		ProjectPath: filepath.Join(dir, "nonexistent.toml"),
	})
	require.NoError(t, err)

	assert.Equal(t, "~/worktrees", cfg.WtBasePath)
	assert.ElementsMatch(t, []SlotDefinition{{Name: "dev"}, {Name: "staging"}}, cfg.Slots)
}

func TestLoadConfig_OnlyProject(t *testing.T) {
	dir := t.TempDir()
	pp := writeTempConfig(t, dir, "project.toml", projectTOML)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  filepath.Join(dir, "nonexistent.toml"),
		ProjectPath: pp,
	})
	require.NoError(t, err)

	assert.Equal(t, "~/project-trees", cfg.WtBasePath)
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

func TestLoadConfig_ProjectOverridesWtBasePath(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
wt_base_path = "~/global-base"

[[slots]]
name = "s1"
`)
	pp := writeTempConfig(t, dir, "project.toml", `
wt_base_path = "~/project-base"

[[slots]]
name = "s1"
`)

	cfg, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.NoError(t, err)
	assert.Equal(t, "~/project-base", cfg.WtBasePath)
}

func TestLoadConfig_ProjectSlotsAppendGlobal(t *testing.T) {
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
	assert.ElementsMatch(t, []SlotDefinition{{Name: "global-slot"}, {Name: "project-slot"}}, cfg.Slots)
}

func TestLoadConfig_HooksFieldMerged(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
[[slots]]
name = "s1"

[hooks]
pre_mount = [{type = "run", command = "global-pre.sh"}]
post_mount = [{type = "run", command = "global-post.sh"}]
`)
	pp := writeTempConfig(t, dir, "project.toml", `
[[slots]]
name = "s1"

[hooks]
post_mount = [{type = "run", command = "project-post.sh"}]
pre_clear = [{type = "run", command = "project-pre-clear.sh"}]
`)

	cfg, err := LoadConfig(LoadOptions{GlobalPath: gp, ProjectPath: pp})
	require.NoError(t, err)
	assert.Equal(t, []HookAction{{Type: "run", Command: "global-pre.sh"}}, cfg.Hooks.PreMount)
	assert.Equal(t, []HookAction{{Type: "run", Command: "project-post.sh"}}, cfg.Hooks.PostMount)
	assert.Equal(t, []HookAction{{Type: "run", Command: "project-pre-clear.sh"}}, cfg.Hooks.PreClear)
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
	require.NoError(t, os.MkdirAll(repoRoot, 0o755))

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

// --- Feature A: .git-slot/config.toml local config support ---

const localTOML = `
wt_base_path = "~/local-trees"

[[slots]]
name = "local-only"

[hooks]
pre_mount = [{type = "run", command = "local-pre.sh"}]
`

// F1-01: 3ファイル全て存在 → local が最高優先でマージ
func TestLoadConfig_F1_01_AllThreeExist(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)
	pp := writeTempConfig(t, dir, "project.toml", projectTOML)
	lp := writeTempConfig(t, dir, "local.toml", localTOML)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  gp,
		ProjectPath: pp,
		LocalPath:   lp,
	})
	require.NoError(t, err)

	// local の wt_base_path が最優先
	assert.Equal(t, "~/local-trees", cfg.WtBasePath)
	// local の hook が最優先
	assert.Equal(t, []HookAction{{Type: "run", Command: "local-pre.sh"}}, cfg.Hooks.PreMount)
	// local のスロットも含まれる
	assert.Contains(t, cfg.Slots, SlotDefinition{Name: "local-only"})
}

// F1-02: global + local のみ（project なし）→ 正常マージ
func TestLoadConfig_F1_02_GlobalAndLocalOnly(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)
	lp := writeTempConfig(t, dir, "local.toml", localTOML)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  gp,
		ProjectPath: filepath.Join(dir, "nonexistent.toml"),
		LocalPath:   lp,
	})
	require.NoError(t, err)

	// local の wt_base_path が上書き
	assert.Equal(t, "~/local-trees", cfg.WtBasePath)
	// global のスロットも含まれる
	assert.Contains(t, cfg.Slots, SlotDefinition{Name: "dev"})
	assert.Contains(t, cfg.Slots, SlotDefinition{Name: "local-only"})
}

// F1-03: local のみ存在 → エラーなし、local の設定返す
func TestLoadConfig_F1_03_LocalOnly(t *testing.T) {
	dir := t.TempDir()
	lp := writeTempConfig(t, dir, "local.toml", localTOML)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  filepath.Join(dir, "nonexistent-global.toml"),
		ProjectPath: filepath.Join(dir, "nonexistent-project.toml"),
		LocalPath:   lp,
	})
	require.NoError(t, err)

	assert.Equal(t, "~/local-trees", cfg.WtBasePath)
	assert.Equal(t, []SlotDefinition{{Name: "local-only"}}, cfg.Slots)
}

// F1-04: 全ファイル不在 → ErrNoConfig
func TestLoadConfig_F1_04_NoneExist(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadConfig(LoadOptions{
		GlobalPath:  filepath.Join(dir, "no-global.toml"),
		ProjectPath: filepath.Join(dir, "no-project.toml"),
		LocalPath:   filepath.Join(dir, "no-local.toml"),
	})
	require.ErrorIs(t, err, ErrNoConfig)
}

// F1-05: local に TOML パースエラー → エラー返す
func TestLoadConfig_F1_05_LocalTOMLParseError(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", globalTOML)
	lp := writeTempConfig(t, dir, "bad-local.toml", `[invalid toml
this is = broken`)

	_, err := LoadConfig(LoadOptions{
		GlobalPath: gp,
		LocalPath:  lp,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigParse)
	assert.Contains(t, err.Error(), "bad-local.toml")
}

// F1-06: LocalPath 未指定・RepoRoot あり → 自動導出して読み込む
func TestLoadConfig_F1_06_LocalPathDerivedFromRepoRoot(t *testing.T) {
	dir := t.TempDir()
	repoRoot := filepath.Join(dir, "repo")
	localConfigDir := filepath.Join(repoRoot, ".git-slot")
	require.NoError(t, os.MkdirAll(localConfigDir, 0o755))

	// .git-slot/config.toml を作成
	err := os.WriteFile(filepath.Join(localConfigDir, "config.toml"), []byte(localTOML), 0o644)
	require.NoError(t, err)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath: filepath.Join(dir, "nonexistent.toml"),
		RepoRoot:   repoRoot,
		// LocalPath は指定しない → 自動導出
	})
	require.NoError(t, err)
	assert.Equal(t, "~/local-trees", cfg.WtBasePath)
	assert.Equal(t, []SlotDefinition{{Name: "local-only"}}, cfg.Slots)
}

// F1-07: 同一 hook type が複数層に存在 → local が上書き
func TestLoadConfig_F1_07_LocalOverridesHooks(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
[[slots]]
name = "s1"

[hooks]
pre_mount = [{type = "run", command = "global-pre.sh"}]
`)
	pp := writeTempConfig(t, dir, "project.toml", `
[[slots]]
name = "s1"

[hooks]
pre_mount = [{type = "run", command = "project-pre.sh"}]
`)
	lp := writeTempConfig(t, dir, "local.toml", `
[[slots]]
name = "s1"

[hooks]
pre_mount = [{type = "run", command = "local-pre.sh"}]
`)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  gp,
		ProjectPath: pp,
		LocalPath:   lp,
	})
	require.NoError(t, err)
	// local の pre_mount が最終的に採用される
	assert.Equal(t, []HookAction{{Type: "run", Command: "local-pre.sh"}}, cfg.Hooks.PreMount)
}

// F1-08: 同名スロットが複数層に存在 → local の値を採用
func TestLoadConfig_F1_08_LocalOverridesSlot(t *testing.T) {
	dir := t.TempDir()
	gp := writeTempConfig(t, dir, "global.toml", `
[[slots]]
name = "shared"
icon = "🌍"
`)
	pp := writeTempConfig(t, dir, "project.toml", `
[[slots]]
name = "shared"
icon = "📁"
`)
	lp := writeTempConfig(t, dir, "local.toml", `
[[slots]]
name = "shared"
icon = "💻"
`)

	cfg, err := LoadConfig(LoadOptions{
		GlobalPath:  gp,
		ProjectPath: pp,
		LocalPath:   lp,
	})
	require.NoError(t, err)
	// local の icon が採用される
	slot := cfg.FindSlot("shared")
	require.NotNil(t, slot)
	assert.Equal(t, "💻", slot.Icon)
}
