package slot

import (
	"testing"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/pathutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlotNameForWorktreeRoot(t *testing.T) {
	cfg := &config.Config{
		Slots: []config.SlotDefinition{
			{Name: "work"},
			{Name: "hotfix"},
		},
	}
	base := "/wt/github.com/o/r"
	repo := "r"
	workPath := pathutil.ResolveSlotPath(base, repo, "work")

	name, err := SlotNameForWorktreeRoot(cfg, base, repo, workPath)
	require.NoError(t, err)
	assert.Equal(t, "work", name)

	_, err = SlotNameForWorktreeRoot(cfg, base, repo, "/somewhere/else")
	assert.ErrorIs(t, err, ErrNotASlotWorktree)

	_, err = SlotNameForWorktreeRoot(nil, base, repo, workPath)
	assert.ErrorIs(t, err, ErrNotASlotWorktree)
}
