package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/AquiTCD/git-slot/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSlotManager struct {
	listFn      func() ([]slot.Slot, error)
	getPathFn   func(string) (string, error)
	slotPathFn  func(string) string
	mountFn     func(string, string, slot.MountOptions) error
	clearFn     func(string, slot.ClearOptions) error
	statusFn    func(string) (*slot.SlotStatus, error)
	statusAllFn func() ([]slot.SlotStatus, error)
}

func (m *mockSlotManager) List() ([]slot.Slot, error) {
	if m.listFn != nil {
		return m.listFn()
	}
	return nil, nil
}

func (m *mockSlotManager) GetPath(name string) (string, error) {
	if m.getPathFn != nil {
		return m.getPathFn(name)
	}
	return "", nil
}

func (m *mockSlotManager) SlotPath(name string) string {
	if m.slotPathFn != nil {
		return m.slotPathFn(name)
	}
	return ""
}

func (m *mockSlotManager) Mount(slotName, branchName string, opts slot.MountOptions) error {
	if m.mountFn != nil {
		return m.mountFn(slotName, branchName, opts)
	}
	return nil
}

func (m *mockSlotManager) Clear(slotName string, opts slot.ClearOptions) error {
	if m.clearFn != nil {
		return m.clearFn(slotName, opts)
	}
	return nil
}

func (m *mockSlotManager) Status(name string) (*slot.SlotStatus, error) {
	if m.statusFn != nil {
		return m.statusFn(name)
	}
	return nil, nil
}

func (m *mockSlotManager) StatusAll() ([]slot.SlotStatus, error) {
	if m.statusAllFn != nil {
		return m.statusAllFn()
	}
	return nil, nil
}

// --- runList tests ---

func TestRunList_Text(t *testing.T) {
	mgr := &mockSlotManager{
		listFn: func() ([]slot.Slot, error) {
			return []slot.Slot{
				{Name: "work", State: slot.SlotActive, Branch: "main"},
				{Name: "hotfix", State: slot.SlotEmpty},
			}, nil
		},
	}

	var buf bytes.Buffer
	err := runList(mgr, &buf, false)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "work")
	assert.Contains(t, buf.String(), "hotfix")
}

func TestRunList_JSON(t *testing.T) {
	mgr := &mockSlotManager{
		listFn: func() ([]slot.Slot, error) {
			return []slot.Slot{
				{Name: "work", State: slot.SlotActive, Branch: "main", Path: "/slots/work"},
			}, nil
		},
	}

	var buf bytes.Buffer
	err := runList(mgr, &buf, true)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"name": "work"`)
	assert.Contains(t, buf.String(), `"state": "active"`)
}

func TestRunList_Error(t *testing.T) {
	mgr := &mockSlotManager{
		listFn: func() ([]slot.Slot, error) {
			return nil, errors.New("list failed")
		},
	}

	var buf bytes.Buffer
	err := runList(mgr, &buf, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list failed")
}

// --- runGetPath tests ---

func TestRunGetPath_Success(t *testing.T) {
	mgr := &mockSlotManager{
		getPathFn: func(name string) (string, error) {
			return "/slots/" + name, nil
		},
	}

	var buf bytes.Buffer
	err := runGetPath(mgr, "work", &buf)
	require.NoError(t, err)
	assert.Equal(t, "/slots/work\n", buf.String())
}

func TestRunGetPath_Error(t *testing.T) {
	mgr := &mockSlotManager{
		getPathFn: func(name string) (string, error) {
			return "", errors.New("not found")
		},
	}

	var buf bytes.Buffer
	err := runGetPath(mgr, "unknown", &buf)
	require.Error(t, err)
}

// --- runStatus tests ---

func TestRunStatus_SingleSlot_Text(t *testing.T) {
	mgr := &mockSlotManager{
		statusFn: func(name string) (*slot.SlotStatus, error) {
			return &slot.SlotStatus{
				Slot: slot.Slot{
					Name:     "work",
					State:    slot.SlotActive,
					Branch:   "main",
					HeadHash: "abc1234",
					Path:     "/slots/work",
				},
				CommitSubject: "initial commit",
			}, nil
		},
	}

	var buf bytes.Buffer
	err := runStatus(mgr, "work", &buf, false)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Slot:    work")
	assert.Contains(t, buf.String(), "Branch:  main")
}

func TestRunStatus_SingleSlot_JSON(t *testing.T) {
	mgr := &mockSlotManager{
		statusFn: func(name string) (*slot.SlotStatus, error) {
			return &slot.SlotStatus{
				Slot: slot.Slot{
					Name:   "work",
					State:  slot.SlotActive,
					Branch: "main",
					Path:   "/slots/work",
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	err := runStatus(mgr, "work", &buf, true)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"name": "work"`)
}

func TestRunStatus_AllSlots_Text(t *testing.T) {
	mgr := &mockSlotManager{
		statusAllFn: func() ([]slot.SlotStatus, error) {
			return []slot.SlotStatus{
				{Slot: slot.Slot{Name: "work", State: slot.SlotActive, Branch: "main", HeadHash: "abc", Path: "/slots/work"}},
				{Slot: slot.Slot{Name: "hotfix", State: slot.SlotEmpty}},
			}, nil
		},
	}

	var buf bytes.Buffer
	err := runStatus(mgr, "", &buf, false)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "work")
	assert.Contains(t, buf.String(), "hotfix")
}

func TestRunStatus_Error(t *testing.T) {
	mgr := &mockSlotManager{
		statusFn: func(name string) (*slot.SlotStatus, error) {
			return nil, errors.New("status failed")
		},
	}

	var buf bytes.Buffer
	err := runStatus(mgr, "work", &buf, false)
	require.Error(t, err)
}

// --- runMount tests ---

func TestRunMount_Success(t *testing.T) {
	mountCalled := false
	mgr := &mockSlotManager{
		mountFn: func(slotName, branchName string, opts slot.MountOptions) error {
			mountCalled = true
			assert.Equal(t, "work", slotName)
			assert.Equal(t, "feature/x", branchName)
			assert.False(t, opts.CreateBranch)
			assert.False(t, opts.Force)
			return nil
		},
		getPathFn: func(name string) (string, error) {
			return "/slots/" + name, nil
		},
	}

	a := &app{
		mgr: mgr,
		cfg: &config.Config{},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runMount(a, "work", "feature/x", mountOptions{}, &buf)
	require.NoError(t, err)
	assert.True(t, mountCalled)
	assert.Contains(t, buf.String(), "/slots/work")
}

func TestRunMount_CreateBranch(t *testing.T) {
	mgr := &mockSlotManager{
		mountFn: func(slotName, branchName string, opts slot.MountOptions) error {
			assert.True(t, opts.CreateBranch)
			return nil
		},
		getPathFn: func(name string) (string, error) {
			return "/slots/" + name, nil
		},
	}

	a := &app{
		mgr: mgr,
		cfg: &config.Config{},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runMount(a, "work", "new-branch", mountOptions{createBranch: true}, &buf)
	require.NoError(t, err)
}

func TestRunMount_Force(t *testing.T) {
	mgr := &mockSlotManager{
		mountFn: func(slotName, branchName string, opts slot.MountOptions) error {
			assert.True(t, opts.Force)
			return nil
		},
		getPathFn: func(name string) (string, error) {
			return "/slots/" + name, nil
		},
	}

	a := &app{
		mgr: mgr,
		cfg: &config.Config{},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runMount(a, "work", "main", mountOptions{force: true}, &buf)
	require.NoError(t, err)
}

func TestRunMount_Error(t *testing.T) {
	mgr := &mockSlotManager{
		mountFn: func(slotName, branchName string, opts slot.MountOptions) error {
			return errors.New("mount failed")
		},
	}

	a := &app{
		mgr: mgr,
		cfg: &config.Config{},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runMount(a, "work", "main", mountOptions{}, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mount failed")
}

// --- runClear tests ---

func TestRunClear_Success(t *testing.T) {
	clearCalled := false
	mgr := &mockSlotManager{
		clearFn: func(name string, opts slot.ClearOptions) error {
			clearCalled = true
			assert.Equal(t, "work", name)
			assert.False(t, opts.Force)
			return nil
		},
	}

	a := &app{
		mgr: mgr,
		cfg: &config.Config{},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runClear(a, "work", false, &buf)
	require.NoError(t, err)
	assert.True(t, clearCalled)
}

func TestRunClear_Force(t *testing.T) {
	mgr := &mockSlotManager{
		clearFn: func(name string, opts slot.ClearOptions) error {
			assert.True(t, opts.Force)
			return nil
		},
	}

	a := &app{
		mgr: mgr,
		cfg: &config.Config{},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runClear(a, "work", true, &buf)
	require.NoError(t, err)
}

func TestRunClear_Error(t *testing.T) {
	mgr := &mockSlotManager{
		clearFn: func(name string, opts slot.ClearOptions) error {
			return errors.New("clear failed")
		},
	}

	a := &app{
		mgr: mgr,
		cfg: &config.Config{},

		repoRoot: "/repo",
	}

	var buf bytes.Buffer
	err := runClear(a, "work", false, &buf)
	require.Error(t, err)
}

// --- buildPostMountHooks tests ---

func TestBuildPostMountHooks_PreservesRunHooks(t *testing.T) {
	existing := []config.HookAction{
		{Type: "run", Command: "echo hello"},
		{Type: "link", Source: "/a", Dest: "/b"},
	}
	items := []tui.HookItem{
		{Path: ".env", Action: tui.ActionNone},
	}

	hooks, hasNew := buildPostMountHooks(existing, items)
	assert.False(t, hasNew)
	assert.Len(t, hooks, 1)
	assert.Equal(t, "run", hooks[0].Type)
}

func TestBuildPostMountHooks_AddsNewItems(t *testing.T) {
	items := []tui.HookItem{
		{Path: ".env", Action: tui.ActionSymlink},
		{Path: "vendor/", Action: tui.ActionCopy},
		{Path: "ignored.txt", Action: tui.ActionNone},
	}

	hooks, hasNew := buildPostMountHooks(nil, items)
	assert.True(t, hasNew)
	assert.Len(t, hooks, 2)
	assert.Equal(t, "link", hooks[0].Type)
	assert.Equal(t, "copy", hooks[1].Type)
}

func TestBuildPostMountHooks_Empty(t *testing.T) {
	hooks, hasNew := buildPostMountHooks(nil, nil)
	assert.False(t, hasNew)
	assert.Empty(t, hooks)
}

// --- generateHooksTOML tests ---

func TestGenerateHooksTOML_WithExistingContent(t *testing.T) {
	original := `[slots]
name = "work"

# --- Managed by git-slot --hook ---
[[hooks.post_mount]]
type = "link"
source = "old"
dest = "old"
`
	hooks := []config.HookAction{
		{Type: "link", Source: "/new/src", Dest: "/new/dest"},
	}

	result, err := generateHooksTOML(original, hooks)
	require.NoError(t, err)

	assert.Contains(t, result, `[slots]`)
	assert.Contains(t, result, hookManagedMarker)
	assert.Contains(t, result, `source = '/new/src'`)
	assert.NotContains(t, result, "old")
}

func TestGenerateHooksTOML_EmptyOriginal(t *testing.T) {
	hooks := []config.HookAction{
		{Type: "copy", Source: "/src", Dest: "/dest"},
	}

	result, err := generateHooksTOML("", hooks)
	require.NoError(t, err)

	assert.Contains(t, result, hookManagedMarker)
	assert.Contains(t, result, "copy")
}

// --- resolveConfigPath tests ---

func TestResolveConfigPath_Project(t *testing.T) {
	path, err := resolveConfigPath("/repo", false)
	require.NoError(t, err)
	assert.Equal(t, "/repo/git-slot.toml", path)
}

func TestResolveConfigPath_Global(t *testing.T) {
	path, err := resolveConfigPath("/repo", true)
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(path, "config.toml"))
	assert.Contains(t, path, "git-slot")
	assert.NotEqual(t, "/repo/git-slot.toml", path)
}

// --- mapExitCode tests ---

func TestMapExitCode_Nil(t *testing.T) {
	assert.Equal(t, 0, mapExitCode(nil))
}

func TestMapExitCode_GenericError(t *testing.T) {
	assert.Equal(t, 1, mapExitCode(errors.New("oops")))
}
