package slot

import (
	"path/filepath"
	"strings"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/pathutil"
)

type SlotManager interface {
	List() ([]Slot, error)
	GetPath(slotName string) (string, error)
	Mount(slotName, branchName string, opts MountOptions) error
	Clear(slotName string, opts ClearOptions) error
	Status(slotName string) (*SlotStatus, error)
	StatusAll() ([]SlotStatus, error)
}

type Manager struct {
	cfg       *config.Config
	basePath  string
	wt        git.Worktree
	worktrees []git.WorktreeInfo
	fetched   bool
}

func NewManager(cfg *config.Config, basePath string, wt git.Worktree) *Manager {
	return &Manager{cfg: cfg, basePath: basePath, wt: wt}
}

func (m *Manager) populateSlot(s *Slot, wtByPath map[string]git.WorktreeInfo) error {
	wt, ok := wtByPath[s.Path]
	if !ok {
		return nil
	}
	s.State = SlotActive
	s.Branch = wt.Branch
	s.HeadHash = wt.HeadHash

	dirty, err := m.wt.IsDirty(s.Path)
	if err != nil {
		return err
	}
	s.IsDirty = dirty
	return nil
}

func (m *Manager) worktreeMap() (map[string]git.WorktreeInfo, error) {
	worktrees, err := m.fetchWorktrees()
	if err != nil {
		return nil, err
	}
	wtByPath := make(map[string]git.WorktreeInfo, len(worktrees))
	for _, wt := range worktrees {
		wtByPath[wt.Path] = wt
	}
	return wtByPath, nil
}

func (m *Manager) List() ([]Slot, error) {
	wtByPath, err := m.worktreeMap()
	if err != nil {
		return nil, err
	}

	slots := make([]Slot, 0, len(m.cfg.Slots))
	for _, def := range m.cfg.Slots {
		s := Slot{
			Name: def.Name,
			Icon: def.Icon,
			Path: pathutil.ResolveSlotPath(m.basePath, def.Name),
		}
		if err := m.populateSlot(&s, wtByPath); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}

	return slots, nil
}

func (m *Manager) GetPath(slotName string) (string, error) {
	slot, err := m.resolveSlot(slotName)
	if err != nil {
		return "", err
	}

	if slot.State == SlotEmpty {
		return "", &SlotError{SlotName: slotName, Err: ErrSlotEmpty, Detail: "mount a branch first"}
	}

	return slot.Path, nil
}

type MountOptions struct {
	CreateBranch bool
	Force        bool
}

func (m *Manager) Mount(slotName, branchName string, opts MountOptions) error {
	slot, err := m.resolveSlot(slotName)
	if err != nil {
		return err
	}

	if slot.State == SlotActive && slot.Branch == branchName {
		return nil
	}

	if slot.State == SlotActive && slot.IsDirty && !opts.Force {
		return &SlotError{SlotName: slotName, Err: ErrSlotDirty}
	}

	worktrees, err := m.fetchWorktrees()
	if err != nil {
		return err
	}
	for _, wt := range worktrees {
		if wt.Path == slot.Path {
			continue
		}
		if wt.Branch == branchName {
			usedBy := m.worktreeLabel(wt.Path)
			return &BranchError{Branch: branchName, Err: ErrBranchInUse, UsedBy: usedBy}
		}
	}

	if opts.CreateBranch {
		exists, err := m.wt.BranchExists(branchName)
		if err != nil {
			return err
		}
		if exists {
			return &BranchError{Branch: branchName, Err: ErrBranchExists}
		}

		if slot.State == SlotActive {
			err = m.wt.SwitchCreate(slot.Path, branchName)
			m.invalidateCache()
			return err
		}
		err = m.wt.AddNewBranch(slot.Path, branchName)
		m.invalidateCache()
		return err
	}

	exists, err := m.wt.BranchExists(branchName)
	if err != nil {
		return err
	}
	if !exists {
		return &BranchError{Branch: branchName, Err: ErrBranchNotFound}
	}

	if slot.State == SlotActive {
		err = m.wt.Switch(slot.Path, branchName, opts.Force)
		m.invalidateCache()
		return err
	}
	err = m.wt.Add(slot.Path, branchName)
	m.invalidateCache()
	return err
}

type ClearOptions struct {
	Force bool
}

func (m *Manager) Clear(slotName string, opts ClearOptions) error {
	slot, err := m.resolveSlot(slotName)
	if err != nil {
		return err
	}

	if slot.State == SlotEmpty {
		return &SlotError{SlotName: slotName, Err: ErrSlotAlreadyEmpty}
	}

	if slot.IsDirty && !opts.Force {
		return &SlotError{SlotName: slotName, Err: ErrSlotDirty}
	}

	err = m.wt.Remove(slot.Path, opts.Force)
	m.invalidateCache()
	return err
}

type SlotStatus struct {
	Slot
	CommitSubject string
	Changes       []string
}

func (m *Manager) enrichStatus(st *SlotStatus) error {
	if st.State != SlotActive {
		return nil
	}
	subject, err := m.wt.CommitSubject(st.Path)
	if err != nil {
		return err
	}
	st.CommitSubject = subject

	changes, err := m.wt.StatusShort(st.Path)
	if err != nil {
		return err
	}
	st.Changes = changes
	return nil
}

func (m *Manager) Status(slotName string) (*SlotStatus, error) {
	s, err := m.resolveSlot(slotName)
	if err != nil {
		return nil, err
	}

	st := &SlotStatus{Slot: *s}
	if err := m.enrichStatus(st); err != nil {
		return nil, err
	}
	return st, nil
}

func (m *Manager) StatusAll() ([]SlotStatus, error) {
	slots, err := m.List()
	if err != nil {
		return nil, err
	}

	statuses := make([]SlotStatus, 0, len(slots))
	for _, s := range slots {
		st := SlotStatus{Slot: s}
		if err := m.enrichStatus(&st); err != nil {
			return nil, err
		}
		statuses = append(statuses, st)
	}

	return statuses, nil
}

func (m *Manager) resolveSlot(name string) (*Slot, error) {
	def := m.cfg.FindSlot(name)
	if def == nil {
		return nil, &SlotError{SlotName: name, Err: ErrSlotUnknown}
	}

	wtByPath, err := m.worktreeMap()
	if err != nil {
		return nil, err
	}

	s := &Slot{
		Name: name,
		Icon: def.Icon,
		Path: pathutil.ResolveSlotPath(m.basePath, name),
	}
	if err := m.populateSlot(s, wtByPath); err != nil {
		return nil, err
	}

	return s, nil
}

// worktreeLabel returns a human-readable label for a worktree path.
// Slot paths under basePath are returned as the slot name; others as the full path.
func (m *Manager) worktreeLabel(wtPath string) string {
	if strings.HasPrefix(wtPath, m.basePath) {
		rel, err := filepath.Rel(m.basePath, wtPath)
		if err == nil && !strings.Contains(rel, string(filepath.Separator)) {
			return rel
		}
	}
	return wtPath
}

func (m *Manager) fetchWorktrees() ([]git.WorktreeInfo, error) {
	if m.fetched {
		return m.worktrees, nil
	}
	wts, err := m.wt.List()
	if err != nil {
		return nil, err
	}
	m.worktrees = wts
	m.fetched = true
	return m.worktrees, nil
}

func (m *Manager) invalidateCache() {
	m.fetched = false
	m.worktrees = nil
}
