package hook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

const DefaultTimeout = 30 * time.Second

var (
	ErrHookNotFound   = errors.New("hook script not found")
	ErrHookPermission = errors.New("hook script is not executable")
	ErrHookTimeout    = errors.New("hook timed out")
	ErrHookFailed     = errors.New("hook failed")
)

type HookEnv struct {
	SlotName string
	SlotPath string
	Branch   string
	RepoRoot string
	Action   string // "load" or "clear"
}

type Runner struct {
	stdout  io.Writer
	stderr  io.Writer
	timeout time.Duration
}

func NewRunner(stdout, stderr io.Writer) *Runner {
	return &Runner{stdout: stdout, stderr: stderr, timeout: DefaultTimeout}
}

func (r *Runner) Run(scriptPath string, env HookEnv) error {
	if scriptPath == "" {
		return nil
	}

	info, err := os.Stat(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: %w", scriptPath, ErrHookNotFound)
		}
		return err
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("%s: %w", scriptPath, ErrHookPermission)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, scriptPath)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	cmd.Env = append(os.Environ(),
		"GSL_SLOT_NAME="+env.SlotName,
		"GSL_SLOT_PATH="+env.SlotPath,
		"GSL_BRANCH="+env.Branch,
		"GSL_REPO_ROOT="+env.RepoRoot,
		"GSL_ACTION="+env.Action,
	)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("%s: %w", scriptPath, ErrHookTimeout)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("%s: exit code %d: %w", scriptPath, exitErr.ExitCode(), ErrHookFailed)
		}
		return err
	}

	return nil
}
