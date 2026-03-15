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

	globalExists := fileExists(globalPath)
	projectExists := projectPath != "" && fileExists(projectPath)

	if !globalExists && !projectExists {
		return nil, ErrNoConfig
	}

	var globalCfg, projectCfg *Config

	if globalExists {
		cfg, err := readAndParse(globalPath)
		if err != nil {
			return nil, err
		}
		globalCfg = cfg
	}

	if projectExists {
		cfg, err := readAndParse(projectPath)
		if err != nil {
			return nil, err
		}
		projectCfg = cfg
	}

	var final *Config
	switch {
	case globalCfg != nil && projectCfg != nil:
		final = Merge(globalCfg, projectCfg)
	case globalCfg != nil:
		final = globalCfg
	default:
		final = projectCfg
	}

	if err := Validate(final); err != nil {
		return nil, err
	}

	return final, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readAndParse(path string) (*Config, error) {
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
