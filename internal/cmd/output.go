package cmd

import (
	"encoding/json"
	"io"

	"github.com/AquiTCD/git-slot/internal/slot"
)

type jsonSlot struct {
	Name        string `json:"name"`
	State       string `json:"state"`
	Branch      string `json:"branch"`
	Head        string `json:"head"`
	Path        string `json:"path"`
	DirtyCount  int    `json:"dirty_count"`
	AheadCount  int    `json:"ahead_count"`
	HasUpstream bool   `json:"has_upstream"`
}

type jsonSlotList struct {
	Slots []jsonSlot `json:"slots"`
}

type jsonSlotStatus struct {
	Name          string   `json:"name"`
	State         string   `json:"state"`
	Branch        string   `json:"branch"`
	Head          string   `json:"head"`
	CommitSubject string   `json:"commit_subject,omitempty"`
	Path          string   `json:"path"`
	Changes       []string `json:"changes,omitempty"`
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func slotToJSON(s slot.Slot) jsonSlot {
	return jsonSlot{
		Name:        s.Name,
		State:       s.DisplayState(),
		Branch:      s.Branch,
		Head:        s.HeadHash,
		Path:        s.Path,
		DirtyCount:  s.DirtyCount,
		AheadCount:  s.AheadCount,
		HasUpstream: s.HasUpstream,
	}
}

func statusToJSON(s slot.SlotStatus) jsonSlotStatus {
	return jsonSlotStatus{
		Name:          s.Name,
		State:         s.DisplayState(),
		Branch:        s.Branch,
		Head:          s.HeadHash,
		CommitSubject: s.CommitSubject,
		Path:          s.Path,
		Changes:       s.Changes,
	}
}
