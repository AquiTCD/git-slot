package slotenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSlotEnv(t *testing.T) {
	t.Run("auto GSL_ vars are always present", func(t *testing.T) {
		env := BuildSlotEnv(SlotInfo{
			SlotName: "main-work",
			SlotPath: "/path/to/slots/main-work",
			Branch:   "feat/nice-ui",
			RepoRoot: "/path/to/repo",
		}, nil)

		assert.Contains(t, env, "GSL_SLOT_NAME=main-work")
		assert.Contains(t, env, "GSL_SLOT_PATH=/path/to/slots/main-work")
		assert.Contains(t, env, "GSL_BRANCH=feat/nice-ui")
		assert.Contains(t, env, "GSL_REPO_ROOT=/path/to/repo")
		assert.Contains(t, env, "GSL_SHELL_SESSION=1")
	})

	t.Run("user-defined env is included", func(t *testing.T) {
		userEnv := map[string]string{
			"PORT":    "3001",
			"APP_ENV": "slot-dev",
		}
		env := BuildSlotEnv(SlotInfo{
			SlotName: "dev",
			SlotPath: "/slots/dev",
			Branch:   "main",
			RepoRoot: "/repo",
		}, userEnv)

		assert.Contains(t, env, "PORT=3001")
		assert.Contains(t, env, "APP_ENV=slot-dev")
	})

	t.Run("user env does not override GSL_ vars", func(t *testing.T) {
		userEnv := map[string]string{
			"GSL_SLOT_NAME": "hacked",
		}
		env := BuildSlotEnv(SlotInfo{
			SlotName: "real",
			SlotPath: "/slots/real",
			Branch:   "main",
			RepoRoot: "/repo",
		}, userEnv)

		assert.Contains(t, env, "GSL_SLOT_NAME=real")
		assert.NotContains(t, env, "GSL_SLOT_NAME=hacked")
	})

	t.Run("nil user env produces only GSL_ vars", func(t *testing.T) {
		env := BuildSlotEnv(SlotInfo{
			SlotName: "s1",
			SlotPath: "/slots/s1",
			Branch:   "main",
			RepoRoot: "/repo",
		}, nil)

		require.Len(t, env, 5)
	})
}

func TestBuildSlotExecEnv(t *testing.T) {
	t.Run("no GSL_SHELL_SESSION", func(t *testing.T) {
		env := BuildSlotExecEnv(SlotInfo{
			SlotName: "main-work",
			SlotPath: "/path/to/slots/main-work",
			Branch:   "feat/x",
			RepoRoot: "/path/to/repo",
		}, nil)

		assert.Contains(t, env, "GSL_SLOT_NAME=main-work")
		assert.Contains(t, env, "GSL_SLOT_PATH=/path/to/slots/main-work")
		assert.Contains(t, env, "GSL_BRANCH=feat/x")
		assert.Contains(t, env, "GSL_REPO_ROOT=/path/to/repo")
		assert.NotContains(t, env, "GSL_SHELL_SESSION")
	})

	t.Run("user env same rules as shell", func(t *testing.T) {
		env := BuildSlotExecEnv(SlotInfo{
			SlotName: "s1",
			SlotPath: "/slots/s1",
			Branch:   "main",
			RepoRoot: "/repo",
		}, map[string]string{"PORT": "3001"})

		assert.Contains(t, env, "PORT=3001")
		require.Len(t, env, 5)
	})
}

func TestMergeEnvWithOS(t *testing.T) {
	osEnv := []string{
		"HOME=/home/user",
		"PATH=/usr/bin",
		"PORT=8080",
	}
	slotEnv := []string{
		"PORT=3001",
		"GSL_SLOT_NAME=dev",
	}

	merged := MergeEnvWithOS(osEnv, slotEnv)

	assert.Contains(t, merged, "HOME=/home/user")
	assert.Contains(t, merged, "PATH=/usr/bin")
	assert.Contains(t, merged, "PORT=3001")
	assert.Contains(t, merged, "GSL_SLOT_NAME=dev")
	assert.NotContains(t, merged, "PORT=8080")
}
