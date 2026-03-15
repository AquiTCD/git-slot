package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AquiTCD/git-slot/internal/config"
	"github.com/AquiTCD/git-slot/internal/git"
)

func runInit(out io.Writer) error {
	detector := git.NewExecDetector("")
	repoRoot, _ := detector.RepoRoot()

	path, err := config.Init(config.InitOptions{
		Global: flagGlobal,
		Force:  flagForce,
	}, repoRoot)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stderr, "Created %s with template configuration.\n", path)
	return nil
}
