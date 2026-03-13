package slot

import (
	"path/filepath"
	"strings"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
	"github.com/AquiTCD/git-slot/internal/pathutil"
)

type Manager struct {
	cfg      *config.Config
	basePath string
	wt       git.Worktree
}

func NewManager(cfg *config.Config, basePath string, wt git.Worktree) *Manager {
	return &Manager{cfg: cfg, basePath: basePath, wt: wt}
}

func (m *Manager) List() ([]Slot, error) {
	worktrees, err := m.wt.List()
	if err != nil {
		return nil, err
	}

	wtByPath := make(map[string]git.WorktreeInfo, len(worktrees))
	for _, wt := range worktrees {
		wtByPath[wt.Path] = wt
	}

	slots := make([]Slot, 0, len(m.cfg.Slots))
	for _, def := range m.cfg.Slots {
		slotPath := pathutil.ResolveSlotPath(m.basePath, def.Name)
		s := Slot{
			Name: def.Name,
			Path: slotPath,
		}

		if wt, ok := wtByPath[slotPath]; ok {
			s.State = SlotActive
			s.Branch = wt.Branch
			s.HeadHash = wt.HeadHash

			dirty, err := m.wt.IsDirty(slotPath)
			if err != nil {
				return nil, err
			}
			s.IsDirty = dirty
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
		return "", &SlotError{SlotName: slotName, Err: ErrSlotEmpty, Detail: "load a branch first"}
	}

	return slot.Path, nil
}

type LoadOptions struct {
	CreateBranch bool
	Force        bool
}

func (m *Manager) Load(slotName, branchName string, opts LoadOptions) error {
	slot, err := m.resolveSlot(slotName)
	if err != nil {
		return err
	}

	if slot.State == SlotActive && slot.IsDirty && !opts.Force {
		return &SlotError{SlotName: slotName, Err: ErrSlotDirty}
	}

	worktrees, err := m.wt.List()
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
			if err := m.wt.Remove(slot.Path, opts.Force); err != nil {
				return err
			}
		}
		return m.wt.AddNewBranch(slot.Path, branchName)
	}

	exists, err := m.wt.BranchExists(branchName)
	if err != nil {
		return err
	}
	if !exists {
		return &BranchError{Branch: branchName, Err: ErrBranchNotFound}
	}

	if slot.State == SlotActive {
		if err := m.wt.Remove(slot.Path, opts.Force); err != nil {
			return err
		}
	}
	return m.wt.Add(slot.Path, branchName)
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

	return m.wt.Remove(slot.Path, opts.Force)
}

func (m *Manager) findSlotDef(name string) bool {
	for _, def := range m.cfg.Slots {
		if def.Name == name {
			return true
		}
	}
	return false
}

func (m *Manager) resolveSlot(name string) (*Slot, error) {
	if !m.findSlotDef(name) {
		return nil, &SlotError{SlotName: name, Err: ErrSlotUnknown}
	}

	slotPath := pathutil.ResolveSlotPath(m.basePath, name)
	s := &Slot{
		Name: name,
		Path: slotPath,
	}

	worktrees, err := m.wt.List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Path == slotPath {
			s.State = SlotActive
			s.Branch = wt.Branch
			s.HeadHash = wt.HeadHash

			dirty, err := m.wt.IsDirty(slotPath)
			if err != nil {
				return nil, err
			}
			s.IsDirty = dirty
			break
		}
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
