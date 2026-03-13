package slot

import (
	"errors"
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWorktree struct {
	listFn         func() ([]git.WorktreeInfo, error)
	addFn          func(path, branch string) error
	addNewBranchFn func(path, newBranch string) error
	removeFn       func(path string, force bool) error
	branchExistsFn func(branch string) (bool, error)
	isDirtyFn      func(path string) (bool, error)
}

func (m *mockWorktree) List() ([]git.WorktreeInfo, error) { return m.listFn() }
func (m *mockWorktree) Add(path, branch string) error     { return m.addFn(path, branch) }
func (m *mockWorktree) AddNewBranch(path, newBranch string) error {
	return m.addNewBranchFn(path, newBranch)
}
func (m *mockWorktree) Remove(path string, force bool) error     { return m.removeFn(path, force) }
func (m *mockWorktree) BranchExists(branch string) (bool, error) { return m.branchExistsFn(branch) }
func (m *mockWorktree) IsDirty(path string) (bool, error)        { return m.isDirtyFn(path) }

var _ git.Worktree = (*mockWorktree)(nil)

// --- List tests ---

func TestList_AllEmpty(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}, {Name: "hotfix"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 2)
	assert.Equal(t, SlotEmpty, slots[0].State)
	assert.Equal(t, SlotEmpty, slots[1].State)
}

func TestList_MixedState(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}, {Name: "hotfix"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "feature/x", HeadHash: "abc1234"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return false, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 2)
	assert.Equal(t, SlotActive, slots[0].State)
	assert.Equal(t, "feature/x", slots[0].Branch)
	assert.Equal(t, "abc1234", slots[0].HeadHash)
	assert.False(t, slots[0].IsDirty)
	assert.Equal(t, SlotEmpty, slots[1].State)
}

func TestList_WithDirtySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "main", HeadHash: "def5678"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return true, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 1)
	assert.True(t, slots[0].IsDirty)
}

func TestList_WorktreeListError(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, errors.New("git failed") },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	_, err := mgr.List()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "git failed")
}

func TestList_OrderMatchesConfig(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "alpha"}, {Name: "beta"}, {Name: "gamma"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 3)
	assert.Equal(t, "alpha", slots[0].Name)
	assert.Equal(t, "beta", slots[1].Name)
	assert.Equal(t, "gamma", slots[2].Name)
}

// --- GetPath tests ---

func TestGetPath_ActiveSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "main"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return false, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	p, err := mgr.GetPath("work")

	require.NoError(t, err)
	assert.Equal(t, "/base/slots/work", p)
}

func TestGetPath_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	_, err := mgr.GetPath("work")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotEmpty))
}

func TestGetPath_UnknownSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{}

	mgr := NewManager(cfg, "/base/slots", mock)
	_, err := mgr.GetPath("nope")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotUnknown))
}

// --- Load tests ---

func TestLoad_ExistingBranch_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var addPath, addBranch string
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(branch string) (bool, error) { return true, nil },
		addFn: func(path, branch string) error {
			addPath = path
			addBranch = branch
			return nil
		},
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/x", LoadOptions{})

	require.NoError(t, err)
	assert.Equal(t, "/base/slots/work", addPath)
	assert.Equal(t, "feature/x", addBranch)
}

func TestLoad_ExistingBranch_ActiveSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var removeCalled bool
	var addCalled bool
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			if removeCalled {
				return nil, nil
			}
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "old-branch"},
			}, nil
		},
		isDirtyFn:      func(path string) (bool, error) { return false, nil },
		branchExistsFn: func(branch string) (bool, error) { return true, nil },
		removeFn: func(path string, force bool) error {
			removeCalled = true
			return nil
		},
		addFn: func(path, branch string) error {
			addCalled = true
			return nil
		},
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/x", LoadOptions{})

	require.NoError(t, err)
	assert.True(t, removeCalled)
	assert.True(t, addCalled)
}

func TestLoad_ExistingBranch_DirtySlot_NoForce(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "old-branch"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return true, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/x", LoadOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotDirty))
}

func TestLoad_ExistingBranch_DirtySlot_Force(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var removeForce bool
	var removeCalled bool
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			if removeCalled {
				return nil, nil
			}
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "old-branch"},
			}, nil
		},
		isDirtyFn:      func(path string) (bool, error) { return true, nil },
		branchExistsFn: func(branch string) (bool, error) { return true, nil },
		removeFn: func(path string, force bool) error {
			removeCalled = true
			removeForce = force
			return nil
		},
		addFn: func(path, branch string) error { return nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/x", LoadOptions{Force: true})

	require.NoError(t, err)
	assert.True(t, removeCalled)
	assert.True(t, removeForce)
}

func TestLoad_CreateBranch_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var newBranchPath, newBranchName string
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(branch string) (bool, error) { return false, nil },
		addNewBranchFn: func(path, newBranch string) error {
			newBranchPath = path
			newBranchName = newBranch
			return nil
		},
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/new", LoadOptions{CreateBranch: true})

	require.NoError(t, err)
	assert.Equal(t, "/base/slots/work", newBranchPath)
	assert.Equal(t, "feature/new", newBranchName)
}

func TestLoad_CreateBranch_AlreadyExists(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(branch string) (bool, error) { return true, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/new", LoadOptions{CreateBranch: true})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchExists))
}

func TestLoad_BranchInUse_OtherSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}, {Name: "hotfix"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/hotfix", Branch: "feature/x"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return false, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/x", LoadOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchInUse))
	var brErr *BranchError
	require.True(t, errors.As(err, &brErr))
	assert.Equal(t, "hotfix", brErr.UsedBy)
}

func TestLoad_BranchInUse_GwqWorktree(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/other/worktree/path", Branch: "feature/x"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return false, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/x", LoadOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchInUse))
	var brErr *BranchError
	require.True(t, errors.As(err, &brErr))
	assert.Equal(t, "/other/worktree/path", brErr.UsedBy)
}

func TestLoad_BranchNotFound(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(branch string) (bool, error) { return false, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "nonexistent", LoadOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchNotFound))
}

func TestLoad_UnknownSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("nope", "main", LoadOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotUnknown))
}

func TestLoad_WorktreeAddFails(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(branch string) (bool, error) { return true, nil },
		addFn: func(path, branch string) error {
			return errors.New("worktree add failed")
		},
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Load("work", "feature/x", LoadOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "worktree add failed")
}

// --- Clear tests ---

func TestClear_ActiveSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var removeCalled bool
	var removePath string
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "main"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return false, nil },
		removeFn: func(path string, force bool) error {
			removeCalled = true
			removePath = path
			return nil
		},
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Clear("work", ClearOptions{})

	require.NoError(t, err)
	assert.True(t, removeCalled)
	assert.Equal(t, "/base/slots/work", removePath)
}

func TestClear_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Clear("work", ClearOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotAlreadyEmpty))
}

func TestClear_DirtySlot_NoForce(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "main"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return true, nil },
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Clear("work", ClearOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotDirty))
}

func TestClear_DirtySlot_Force(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var removeForce bool
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "main"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return true, nil },
		removeFn: func(path string, force bool) error {
			removeForce = force
			return nil
		},
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Clear("work", ClearOptions{Force: true})

	require.NoError(t, err)
	assert.True(t, removeForce)
}

func TestClear_UnknownSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Clear("nope", ClearOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotUnknown))
}

func TestClear_RemoveFails(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/slots/work", Branch: "main"},
			}, nil
		},
		isDirtyFn: func(path string) (bool, error) { return false, nil },
		removeFn: func(path string, force bool) error {
			return errors.New("remove failed")
		},
	}

	mgr := NewManager(cfg, "/base/slots", mock)
	err := mgr.Clear("work", ClearOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove failed")
}
