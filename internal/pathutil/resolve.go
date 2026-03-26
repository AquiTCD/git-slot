package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
)

const DefaultGwqBaseDir = "~/worktrees"

var ErrNoRemoteInfo = errors.New("remote info is required to resolve gwq-compliant path")

func ResolveSlotsBasePath(cfg *config.Config, remote *git.RemoteInfo) (string, error) {
	if remote == nil {
		return "", ErrNoRemoteInfo
	}

	base := cfg.WtBasePath
	if base == "" {
		base = DefaultGwqBaseDir
	}

	expanded, err := ExpandHome(base)
	if err != nil {
		return "", err
	}

	return filepath.Join(expanded, remote.Host, remote.Owner, remote.Repo), nil
}

func ResolveSlotPath(basePath, slotName string) string {
	return filepath.Join(basePath, slotName)
}

func ExpandHome(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir()
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}
