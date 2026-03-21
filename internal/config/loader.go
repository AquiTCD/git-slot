package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AquiTCD/git-slot/internal/errutil"
)

var ErrNoConfig = errutil.NewExitError("no configuration file found; run `git slot --init` to create one", 2)

type LoadOptions struct {
	GlobalPath  string
	ProjectPath string
	LocalPath   string
	RepoRoot    string
}

func DefaultGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "git-slot", "config.toml"), nil
}

func LoadConfig(opts LoadOptions) (*Config, error) {
	globalPath := opts.GlobalPath
	if globalPath == "" {
		p, err := DefaultGlobalConfigPath()
		if err != nil {
			return nil, err
		}
		globalPath = p
	}

	projectPath := opts.ProjectPath
	if projectPath == "" && opts.RepoRoot != "" {
		projectPath = filepath.Join(opts.RepoRoot, "git-slot.toml")
	}

	localPath := opts.LocalPath
	if localPath == "" && opts.RepoRoot != "" {
		localPath = filepath.Join(opts.RepoRoot, ".git-slot", "config.toml")
	}

	globalExists := fileExists(globalPath)
	projectExists := projectPath != "" && fileExists(projectPath)
	localExists := localPath != "" && fileExists(localPath)

	if !globalExists && !projectExists && !localExists {
		return nil, ErrNoConfig
	}

	var globalCfg, projectCfg, localCfg *Config

	if globalExists {
		cfg, err := LoadSpecificConfig(globalPath)
		if err != nil {
			return nil, err
		}
		globalCfg = cfg
	}

	if projectExists {
		cfg, err := LoadSpecificConfig(projectPath)
		if err != nil {
			return nil, err
		}
		projectCfg = cfg
	}

	if localExists {
		cfg, err := LoadSpecificConfig(localPath)
		if err != nil {
			return nil, err
		}
		localCfg = cfg
	}

	// Merge(Merge(global, project), local)
	intermediate := Merge(globalCfg, projectCfg)
	final := Merge(intermediate, localCfg)

	if err := Validate(final); err != nil {
		return nil, err
	}

	return final, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func LoadSpecificConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	cfg, err := ParseTOML(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}
