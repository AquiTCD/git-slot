package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlotToJSON(t *testing.T) {
	tests := []struct {
		name string
		slot slot.Slot
		want jsonSlot
	}{
		{
			name: "active slot",
			slot: slot.Slot{
				Name:     "main-work",
				Path:     "/home/user/slots/main-work",
				State:    slot.SlotActive,
				Branch:   "feature/nice-ui",
				HeadHash: "a1b2c3d",
			},
			want: jsonSlot{
				Name:   "main-work",
				State:  "active",
				Branch: "feature/nice-ui",
				Head:   "a1b2c3d",
				Path:   "/home/user/slots/main-work",
			},
		},
		{
			name: "dirty slot",
			slot: slot.Slot{
				Name:       "dev",
				Path:       "/home/user/slots/dev",
				State:      slot.SlotActive,
				Branch:     "develop",
				HeadHash:   "f00baa",
				DirtyCount: 1,
			},
			want: jsonSlot{
				Name:       "dev",
				State:      "dirty",
				Branch:     "develop",
				Head:       "f00baa",
				Path:       "/home/user/slots/dev",
				DirtyCount: 1,
			},
		},
		{
			name: "empty slot",
			slot: slot.Slot{
				Name:  "experiment",
				Path:  "/home/user/slots/experiment",
				State: slot.SlotEmpty,
			},
			want: jsonSlot{
				Name:  "experiment",
				State: "empty",
				Path:  "/home/user/slots/experiment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slotToJSON(tt.slot)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatusToJSON(t *testing.T) {
	tests := []struct {
		name   string
		status slot.SlotStatus
		want   jsonSlotStatus
	}{
		{
			name: "active with changes",
			status: slot.SlotStatus{
				Slot: slot.Slot{
					Name:       "main-work",
					Path:       "/home/user/slots/main-work",
					State:      slot.SlotActive,
					Branch:     "feature/nice-ui",
					HeadHash:   "a1b2c3d",
					DirtyCount: 2,
				},
				CommitSubject: "feat: add nice UI component",
				Changes:       []string{"M  src/components/Button.tsx", "?? src/components/Modal.tsx"},
			},
			want: jsonSlotStatus{
				Name:          "main-work",
				State:         "dirty",
				Branch:        "feature/nice-ui",
				Head:          "a1b2c3d",
				CommitSubject: "feat: add nice UI component",
				Path:          "/home/user/slots/main-work",
				Changes:       []string{"M  src/components/Button.tsx", "?? src/components/Modal.tsx"},
			},
		},
		{
			name: "empty slot status",
			status: slot.SlotStatus{
				Slot: slot.Slot{
					Name:  "experiment",
					Path:  "/home/user/slots/experiment",
					State: slot.SlotEmpty,
				},
			},
			want: jsonSlotStatus{
				Name:  "experiment",
				State: "empty",
				Path:  "/home/user/slots/experiment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusToJSON(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWriteJSON_SlotList(t *testing.T) {
	slots := []slot.Slot{
		{
			Name:       "main-work",
			Path:       "/home/user/slots/main-work",
			State:      slot.SlotActive,
			Branch:     "feature/nice-ui",
			HeadHash:   "a1b2c3d",
			DirtyCount: 1,
		},
		{
			Name:  "experiment",
			Path:  "/home/user/slots/experiment",
			State: slot.SlotEmpty,
		},
	}

	items := make([]jsonSlot, len(slots))
	for i, s := range slots {
		items[i] = slotToJSON(s)
	}

	var buf bytes.Buffer
	err := writeJSON(&buf, jsonSlotList{Slots: items})
	require.NoError(t, err)

	var result jsonSlotList
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	require.Len(t, result.Slots, 2)
	assert.Equal(t, "main-work", result.Slots[0].Name)
	assert.Equal(t, "dirty", result.Slots[0].State)
	assert.Equal(t, "feature/nice-ui", result.Slots[0].Branch)
	assert.Equal(t, "a1b2c3d", result.Slots[0].Head)
	assert.Equal(t, "/home/user/slots/main-work", result.Slots[0].Path)

	assert.Equal(t, "experiment", result.Slots[1].Name)
	assert.Equal(t, "empty", result.Slots[1].State)
	assert.Equal(t, "", result.Slots[1].Branch)
	assert.Equal(t, "", result.Slots[1].Head)
	assert.Equal(t, "/home/user/slots/experiment", result.Slots[1].Path)
}

func TestWriteJSON_Status(t *testing.T) {
	st := slot.SlotStatus{
		Slot: slot.Slot{
			Name:       "main-work",
			Path:       "/home/user/slots/main-work",
			State:      slot.SlotActive,
			Branch:     "feature/nice-ui",
			HeadHash:   "a1b2c3d",
			DirtyCount: 1,
		},
		CommitSubject: "feat: add nice UI component",
		Changes:       []string{"M  src/components/Button.tsx", "?? src/components/Modal.tsx"},
	}

	var buf bytes.Buffer
	err := writeJSON(&buf, statusToJSON(st))
	require.NoError(t, err)

	var result jsonSlotStatus
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	assert.Equal(t, "main-work", result.Name)
	assert.Equal(t, "dirty", result.State)
	assert.Equal(t, "feature/nice-ui", result.Branch)
	assert.Equal(t, "a1b2c3d", result.Head)
	assert.Equal(t, "feat: add nice UI component", result.CommitSubject)
	assert.Equal(t, "/home/user/slots/main-work", result.Path)
	assert.Equal(t, []string{"M  src/components/Button.tsx", "?? src/components/Modal.tsx"}, result.Changes)
}

func TestWriteJSON_StatusAll(t *testing.T) {
	statuses := []slot.SlotStatus{
		{
			Slot: slot.Slot{
				Name:     "main-work",
				Path:     "/home/user/slots/main-work",
				State:    slot.SlotActive,
				Branch:   "feature/nice-ui",
				HeadHash: "a1b2c3d",
			},
			CommitSubject: "feat: add nice UI component",
		},
		{
			Slot: slot.Slot{
				Name:  "experiment",
				Path:  "/home/user/slots/experiment",
				State: slot.SlotEmpty,
			},
		},
	}

	items := make([]jsonSlotStatus, len(statuses))
	for i, s := range statuses {
		items[i] = statusToJSON(s)
	}

	var buf bytes.Buffer
	err := writeJSON(&buf, items)
	require.NoError(t, err)

	var result []jsonSlotStatus
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	require.Len(t, result, 2)

	assert.Equal(t, "main-work", result[0].Name)
	assert.Equal(t, "active", result[0].State)
	assert.Equal(t, "feat: add nice UI component", result[0].CommitSubject)
	assert.Empty(t, result[0].Changes)

	assert.Equal(t, "experiment", result[1].Name)
	assert.Equal(t, "empty", result[1].State)
	assert.Empty(t, result[1].CommitSubject)
	assert.Empty(t, result[1].Changes)
}

func TestWriteJSON_EmptySlotList(t *testing.T) {
	var buf bytes.Buffer
	err := writeJSON(&buf, jsonSlotList{Slots: []jsonSlot{}})
	require.NoError(t, err)

	var result jsonSlotList
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Empty(t, result.Slots)
}

func TestWriteJSON_OmitsEmptyOptionalFields(t *testing.T) {
	st := statusToJSON(slot.SlotStatus{
		Slot: slot.Slot{
			Name:  "empty-slot",
			Path:  "/home/user/slots/empty-slot",
			State: slot.SlotEmpty,
		},
	})

	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, st))

	var raw map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &raw))

	assert.Equal(t, "empty-slot", raw["name"])
	assert.Equal(t, "empty", raw["state"])
	assert.Equal(t, "", raw["branch"])
	assert.Equal(t, "", raw["head"])
	assert.Equal(t, "/home/user/slots/empty-slot", raw["path"])

	_, hasCommitSubject := raw["commit_subject"]
	assert.False(t, hasCommitSubject, "commit_subject should be omitted when empty")

	_, hasChanges := raw["changes"]
	assert.False(t, hasChanges, "changes should be omitted when empty")
}

// F2-J1: slotToJSON includes dirty_count and ahead_count fields
func TestSlotToJSON_RichFields_DirtyAndAhead(t *testing.T) {
	s := slot.Slot{
		Name:        "work",
		Path:        "/slots/work",
		State:       slot.SlotActive,
		Branch:      "feat/auth",
		HeadHash:    "abc1234",
		DirtyCount:  2,
		HasUpstream: true,
		AheadCount:  1,
	}
	got := slotToJSON(s)
	assert.Equal(t, 2, got.DirtyCount)
	assert.Equal(t, 1, got.AheadCount)
	assert.True(t, got.HasUpstream)
}

// F2-J2: clean slot with upstream has dirty_count=0 and ahead_count=0 in JSON
func TestSlotToJSON_RichFields_Clean(t *testing.T) {
	s := slot.Slot{
		Name:        "work",
		Path:        "/slots/work",
		State:       slot.SlotActive,
		Branch:      "main",
		HeadHash:    "abc1234",
		DirtyCount:  0,
		HasUpstream: true,
		AheadCount:  0,
	}

	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, slotToJSON(s)))

	var raw map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &raw))

	dirtyCount, ok := raw["dirty_count"]
	assert.True(t, ok, "dirty_count should be present even when 0")
	assert.Equal(t, float64(0), dirtyCount)

	aheadCount, ok := raw["ahead_count"]
	assert.True(t, ok, "ahead_count should be present even when 0")
	assert.Equal(t, float64(0), aheadCount)

	hasUpstream, ok := raw["has_upstream"]
	assert.True(t, ok, "has_upstream should be present")
	assert.Equal(t, true, hasUpstream)
}

func TestWriteJSON_PrettyPrinted(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, jsonSlotList{Slots: []jsonSlot{
		{Name: "a", State: "empty", Path: "/p"},
	}}))

	output := buf.String()
	assert.Contains(t, output, "  ")
	assert.Contains(t, output, "\n")
}

// ENH-J1: statusToJSONSlot includes behind_count and commit_subject
func TestStatusToJSONSlot_BehindAndCommitSubject(t *testing.T) {
	s := slot.SlotStatus{
		Slot: slot.Slot{
			Name:        "work",
			Path:        "/slots/work",
			State:       slot.SlotActive,
			Branch:      "feat/auth",
			HeadHash:    "abc1234",
			HasUpstream: true,
			AheadCount:  1,
			BehindCount: 2,
		},
		CommitSubject: "feat: add auth",
	}
	got := statusToJSONSlot(s)
	assert.Equal(t, 2, got.BehindCount)
	assert.Equal(t, "feat: add auth", got.CommitSubject)
	assert.Equal(t, 1, got.AheadCount)
	assert.True(t, got.HasUpstream)
}

// ENH-J2: empty slot statusToJSONSlot omits commit_subject
func TestStatusToJSONSlot_EmptySlot_OmitsCommitSubject(t *testing.T) {
	s := slot.SlotStatus{
		Slot: slot.Slot{
			Name:  "experiment",
			Path:  "/slots/experiment",
			State: slot.SlotEmpty,
		},
	}

	var buf bytes.Buffer
	require.NoError(t, writeJSON(&buf, statusToJSONSlot(s)))

	var raw map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &raw))

	_, hasCommitSubject := raw["commit_subject"]
	assert.False(t, hasCommitSubject, "commit_subject should be omitted for empty slot")
	assert.Equal(t, float64(0), raw["behind_count"])
}
