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
	listFn          func() ([]git.WorktreeInfo, error)
	addFn           func(path, branch string) error
	addNewBranchFn  func(path, newBranch string) error
	removeFn        func(path string, force bool) error
	switchFn        func(path, branch string, discardChanges bool) error
	switchCreateFn  func(path, newBranch string) error
	branchExistsFn  func(branch string) (bool, error)
	dirtyCountFn    func(path string) (int, error)
	aheadCountFn    func(path string) (int, error)
	commitSubjectFn func(path string) (string, error)
	statusShortFn   func(path string) ([]string, error)
	listBranchesFn  func() ([]string, error)
}

func (m *mockWorktree) List() ([]git.WorktreeInfo, error) { return m.listFn() }
func (m *mockWorktree) Add(path, branch string) error     { return m.addFn(path, branch) }
func (m *mockWorktree) AddNewBranch(path, newBranch string) error {
	return m.addNewBranchFn(path, newBranch)
}
func (m *mockWorktree) Remove(path string, force bool) error { return m.removeFn(path, force) }
func (m *mockWorktree) Switch(path, branch string, discardChanges bool) error {
	return m.switchFn(path, branch, discardChanges)
}
func (m *mockWorktree) SwitchCreate(path, newBranch string) error {
	return m.switchCreateFn(path, newBranch)
}
func (m *mockWorktree) BranchExists(branch string) (bool, error) { return m.branchExistsFn(branch) }
func (m *mockWorktree) IsDirty(path string) (bool, error) {
	count, err := m.DirtyCount(path)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func (m *mockWorktree) DirtyCount(path string) (int, error) {
	if m.dirtyCountFn != nil {
		return m.dirtyCountFn(path)
	}
	return 0, nil
}
func (m *mockWorktree) AheadCount(path string) (int, error) {
	if m.aheadCountFn != nil {
		return m.aheadCountFn(path)
	}
	return 0, nil
}
func (m *mockWorktree) CommitSubject(path string) (string, error) {
	return m.commitSubjectFn(path)
}
func (m *mockWorktree) StatusShort(path string) ([]string, error) {
	return m.statusShortFn(path)
}
func (m *mockWorktree) ListBranches() ([]string, error) {
	if m.listBranchesFn != nil {
		return m.listBranchesFn()
	}
	return nil, nil
}

var _ git.Worktree = (*mockWorktree)(nil)

// --- List tests ---

func TestList_AllEmpty(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}, {Name: "hotfix"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
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
				{Path: "/base/repo@work", Branch: "feature/x", HeadHash: "abc1234"},
			}, nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 2)
	assert.Equal(t, SlotActive, slots[0].State)
	assert.Equal(t, "feature/x", slots[0].Branch)
	assert.Equal(t, "abc1234", slots[0].HeadHash)
	assert.Equal(t, 0, slots[0].DirtyCount)
	assert.Equal(t, SlotEmpty, slots[1].State)
}

func TestList_WithDirtySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "main", HeadHash: "def5678"},
			}, nil
		},

		dirtyCountFn: func(_ string) (int, error) { return 1, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 1)
	assert.Greater(t, slots[0].DirtyCount, 0)
}

func TestList_WorktreeListError(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, errors.New("git failed") },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
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

	mgr := NewManager(cfg, "/base", "repo", mock)
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
				{Path: "/base/repo@work", Branch: "main"},
			}, nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	p, err := mgr.GetPath("work")

	require.NoError(t, err)
	assert.Equal(t, "/base/repo@work", p)
}

func TestGetPath_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	_, err := mgr.GetPath("work")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotEmpty))
}

func TestGetPath_UnknownSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{}

	mgr := NewManager(cfg, "/base", "repo", mock)
	_, err := mgr.GetPath("nope")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotUnknown))
}

// --- Mount tests ---

func TestMount_ExistingBranch_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var addPath, addBranch string
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(_ string) (bool, error) { return true, nil },
		addFn: func(path, branch string) error {
			addPath = path
			addBranch = branch
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

	require.NoError(t, err)
	assert.Equal(t, "/base/repo@work", addPath)
	assert.Equal(t, "feature/x", addBranch)
}

func TestMount_ExistingBranch_ActiveSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var switchCalled bool
	var switchBranch string
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "old-branch"},
			}, nil
		},

		branchExistsFn: func(_ string) (bool, error) { return true, nil },
		switchFn: func(path, branch string, discard bool) error {
			switchCalled = true
			switchBranch = branch
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

	require.NoError(t, err)
	assert.True(t, switchCalled, "should use switch instead of remove+add")
	assert.Equal(t, "feature/x", switchBranch)
}

func TestMount_SameBranch_Noop(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var anyCalled bool
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "feature/x"},
			}, nil
		},

		switchFn: func(path, branch string, discard bool) error {
			anyCalled = true
			return nil
		},
		addFn: func(path, branch string) error {
			anyCalled = true
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

	require.NoError(t, err)
	assert.False(t, anyCalled, "same branch should be a no-op")
}

func TestMount_DirtySameBranch_ReturnsNil(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var anyCalled bool
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "feature/x"},
			}, nil
		},

		dirtyCountFn: func(_ string) (int, error) { return 1, nil },
		switchFn: func(path, branch string, discard bool) error {
			anyCalled = true
			return nil
		},
		addFn: func(path, branch string) error {
			anyCalled = true
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

	require.NoError(t, err, "dirty slot already on requested branch should be a no-op")
	assert.False(t, anyCalled, "no switch or add should be called for same-branch no-op")
}

func TestMount_CreateBranch_ActiveSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var switchCreateCalled bool
	var switchCreateBranch string
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "old-branch"},
			}, nil
		},

		branchExistsFn: func(_ string) (bool, error) { return false, nil },
		switchCreateFn: func(path, newBranch string) error {
			switchCreateCalled = true
			switchCreateBranch = newBranch
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/new", MountOptions{CreateBranch: true})

	require.NoError(t, err)
	assert.True(t, switchCreateCalled, "should use switchCreate for active slot")
	assert.Equal(t, "feature/new", switchCreateBranch)
}

func TestMount_ExistingBranch_DirtySlot_NoForce(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "old-branch"},
			}, nil
		},

		dirtyCountFn: func(_ string) (int, error) { return 1, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotDirty))
}

func TestMount_ExistingBranch_DirtySlot_Force(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var switchDiscard bool
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "old-branch"},
			}, nil
		},

		dirtyCountFn:   func(_ string) (int, error) { return 1, nil },
		branchExistsFn: func(_ string) (bool, error) { return true, nil },
		switchFn: func(path, branch string, discard bool) error {
			switchDiscard = discard
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{Force: true})

	require.NoError(t, err)
	assert.True(t, switchDiscard, "force should pass discardChanges=true to switch")
}

func TestMount_CreateBranch_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	var newBranchPath, newBranchName string
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(_ string) (bool, error) { return false, nil },
		addNewBranchFn: func(path, newBranch string) error {
			newBranchPath = path
			newBranchName = newBranch
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/new", MountOptions{CreateBranch: true})

	require.NoError(t, err)
	assert.Equal(t, "/base/repo@work", newBranchPath)
	assert.Equal(t, "feature/new", newBranchName)
}

func TestMount_CreateBranch_AlreadyExists(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(_ string) (bool, error) { return true, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/new", MountOptions{CreateBranch: true})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchExists))
}

func TestMount_BranchInUse_OtherSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}, {Name: "hotfix"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@hotfix", Branch: "feature/x"},
			}, nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchInUse))
	var brErr *BranchError
	require.True(t, errors.As(err, &brErr))
	assert.Equal(t, "repo@hotfix", brErr.UsedBy)
}

func TestMount_BranchInUse_GwqWorktree(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/other/worktree/path", Branch: "feature/x"},
			}, nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchInUse))
	var brErr *BranchError
	require.True(t, errors.As(err, &brErr))
	assert.Equal(t, "/other/worktree/path", brErr.UsedBy)
}

func TestMount_BranchNotFound(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(_ string) (bool, error) { return false, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "nonexistent", MountOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrBranchNotFound))
}

func TestMount_UnknownSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("nope", "main", MountOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSlotUnknown))
}

func TestMount_WorktreeAddFails(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn:         func() ([]git.WorktreeInfo, error) { return nil, nil },
		branchExistsFn: func(_ string) (bool, error) { return true, nil },
		addFn: func(path, branch string) error {
			return errors.New("worktree add failed")
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Mount("work", "feature/x", MountOptions{})

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
				{Path: "/base/repo@work", Branch: "main"},
			}, nil
		},

		removeFn: func(path string, _ bool) error {
			removeCalled = true
			removePath = path
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Clear("work", ClearOptions{})

	require.NoError(t, err)
	assert.True(t, removeCalled)
	assert.Equal(t, "/base/repo@work", removePath)
}

func TestClear_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
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
				{Path: "/base/repo@work", Branch: "main"},
			}, nil
		},

		dirtyCountFn: func(_ string) (int, error) { return 1, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
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
				{Path: "/base/repo@work", Branch: "main"},
			}, nil
		},

		dirtyCountFn: func(_ string) (int, error) { return 1, nil },
		removeFn: func(path string, force bool) error {
			removeForce = force
			return nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Clear("work", ClearOptions{Force: true})

	require.NoError(t, err)
	assert.True(t, removeForce)
}

func TestClear_UnknownSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{}

	mgr := NewManager(cfg, "/base", "repo", mock)
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
				{Path: "/base/repo@work", Branch: "main"},
			}, nil
		},

		removeFn: func(path string, force bool) error {
			return errors.New("remove failed")
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	err := mgr.Clear("work", ClearOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove failed")
}

// --- Status tests ---

func TestStatus_ActiveSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "feature/x", HeadHash: "abc1234"},
			}, nil
		},

		dirtyCountFn:    func(_ string) (int, error) { return 2, nil },
		commitSubjectFn: func(path string) (string, error) { return "add feature X", nil },
		statusShortFn: func(path string) ([]string, error) {
			return []string{" M file.go", "?? new.txt"}, nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	st, err := mgr.Status("work")

	require.NoError(t, err)
	assert.Equal(t, "work", st.Name)
	assert.Equal(t, SlotActive, st.State)
	assert.Equal(t, "feature/x", st.Branch)
	assert.Equal(t, "add feature X", st.CommitSubject)
	assert.Equal(t, []string{" M file.go", "?? new.txt"}, st.Changes)
}

func TestStatus_EmptySlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	st, err := mgr.Status("work")

	require.NoError(t, err)
	assert.Equal(t, SlotEmpty, st.State)
	assert.Empty(t, st.CommitSubject)
	assert.Nil(t, st.Changes)
}

func TestStatus_UnknownSlot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{}

	mgr := NewManager(cfg, "/base", "repo", mock)
	_, err := mgr.Status("nope")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSlotUnknown)
}

// --- List rich fields tests (Feature B) ---

// F2-M1: dirty 3 files + ahead 2 commits are reflected in Slot fields
func TestList_RichFields_DirtyAndAhead(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "feature/x", HeadHash: "abc1234"},
			}, nil
		},

		dirtyCountFn: func(_ string) (int, error) { return 3, nil },
		aheadCountFn: func(_ string) (int, error) { return 2, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 1)
	assert.Equal(t, 3, slots[0].DirtyCount)
	assert.Equal(t, 2, slots[0].AheadCount)
	assert.True(t, slots[0].HasUpstream)
}

// F2-M2: upstream not configured => AheadCount=0, HasUpstream=false
func TestList_RichFields_NoUpstream(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "feature/x", HeadHash: "abc1234"},
			}, nil
		},
		dirtyCountFn: func(_ string) (int, error) { return 0, nil },
		aheadCountFn: func(_ string) (int, error) { return 0, errors.New("no upstream") },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 1)
	assert.Equal(t, 0, slots[0].AheadCount)
	assert.False(t, slots[0].HasUpstream)
}

// F2-M3: empty slot => git commands not called
func TestList_RichFields_EmptySlot_NoGitCalls(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	dirtyCountCalled := false
	aheadCountCalled := false
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) { return nil, nil },
		dirtyCountFn: func(_ string) (int, error) {
			dirtyCountCalled = true
			return 0, nil
		},
		aheadCountFn: func(_ string) (int, error) {
			aheadCountCalled = true
			return 0, nil
		},
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	slots, err := mgr.List()

	require.NoError(t, err)
	require.Len(t, slots, 1)
	assert.Equal(t, SlotEmpty, slots[0].State)
	assert.False(t, dirtyCountCalled, "DirtyCount should not be called for empty slot")
	assert.False(t, aheadCountCalled, "AheadCount should not be called for empty slot")
}

// F2-M4: DirtyCount git error propagates
func TestList_RichFields_DirtyCountError(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "feature/x", HeadHash: "abc1234"},
			}, nil
		},
		dirtyCountFn: func(_ string) (int, error) { return 0, errors.New("git status failed") },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	_, err := mgr.List()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "git status failed")
}

func TestStatusAll(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{{Name: "work"}, {Name: "hotfix"}},
	}
	mock := &mockWorktree{
		listFn: func() ([]git.WorktreeInfo, error) {
			return []git.WorktreeInfo{
				{Path: "/base/repo@work", Branch: "feature/x", HeadHash: "abc1234"},
			}, nil
		},

		commitSubjectFn: func(path string) (string, error) { return "initial commit", nil },
		statusShortFn:   func(path string) ([]string, error) { return nil, nil },
	}

	mgr := NewManager(cfg, "/base", "repo", mock)
	statuses, err := mgr.StatusAll()

	require.NoError(t, err)
	require.Len(t, statuses, 2)

	assert.Equal(t, "work", statuses[0].Name)
	assert.Equal(t, SlotActive, statuses[0].State)
	assert.Equal(t, "initial commit", statuses[0].CommitSubject)

	assert.Equal(t, "hotfix", statuses[1].Name)
	assert.Equal(t, SlotEmpty, statuses[1].State)
	assert.Empty(t, statuses[1].CommitSubject)
}
