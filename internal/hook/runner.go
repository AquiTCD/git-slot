package hook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AquiTCD/git-slot/internal/config"
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
	Action   string // "mount" or "clear"
	UserEnv  map[string]string
}

type Runner struct {
	stdout  io.Writer
	stderr  io.Writer
	timeout time.Duration
}

func NewRunner(stdout, stderr io.Writer) *Runner {
	return &Runner{stdout: stdout, stderr: stderr, timeout: DefaultTimeout}
}

func (r *Runner) Run(actions []config.HookAction, env HookEnv) error {
	for _, action := range actions {
		if err := r.runAction(action, env); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runAction(action config.HookAction, env HookEnv) error {
	switch action.Type {
	case "link":
		return r.handleLink(action, env)
	case "copy":
		return r.handleCopy(action, env)
	case "run":
		return r.handleRun(action, env)
	default:
		if action.Command != "" {
			return r.handleRun(action, env) // Default to run for backward compatibility or simple entries
		}
		return nil
	}
}

func (r *Runner) resolveAndPrepare(action config.HookAction, env HookEnv, hookType string) (src, dest string, err error) {
	src = r.expandEnv(action.Source, env)
	dest = r.expandEnv(action.Dest, env)

	if src == "" || dest == "" {
		return "", "", fmt.Errorf("%s hook: missing source or destination", hookType)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", "", err
	}

	_ = os.RemoveAll(dest)
	return src, dest, nil
}

func (r *Runner) handleLink(action config.HookAction, env HookEnv) error {
	src, dest, err := r.resolveAndPrepare(action, env, "link")
	if err != nil {
		return err
	}
	return os.Symlink(src, dest)
}

func (r *Runner) handleCopy(action config.HookAction, env HookEnv) error {
	src, dest, err := r.resolveAndPrepare(action, env, "copy")
	if err != nil {
		return err
	}
	cmd := exec.Command("cp", "-R", src, dest)
	return cmd.Run()
}

func (r *Runner) handleRun(action config.HookAction, env HookEnv) error {
	command := r.expandEnv(action.Command, env)
	if command == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Use shell to execute the command to support pipes, redirection etc.
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	hookEnvVars := []string{
		"GSL_SLOT_NAME=" + env.SlotName,
		"GSL_SLOT_PATH=" + env.SlotPath,
		"GSL_BRANCH=" + env.Branch,
		"GSL_REPO_ROOT=" + env.RepoRoot,
		"GSL_ACTION=" + env.Action,
	}
	for k, v := range env.UserEnv {
		hookEnvVars = append(hookEnvVars, k+"="+v)
	}
	cmd.Env = append(os.Environ(), hookEnvVars...)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command '%s': %w", command, ErrHookTimeout)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command '%s' failed with exit code %d: %w", command, exitErr.ExitCode(), ErrHookFailed)
		}
		return fmt.Errorf("command '%s': %w", command, ErrHookFailed)
	}

	return nil
}

func (r *Runner) expandEnv(s string, env HookEnv) string {
	s = strings.ReplaceAll(s, "$GSL_SLOT_NAME", env.SlotName)
	s = strings.ReplaceAll(s, "$GSL_SLOT_PATH", env.SlotPath)
	s = strings.ReplaceAll(s, "$GSL_BRANCH", env.Branch)
	s = strings.ReplaceAll(s, "$GSL_REPO_ROOT", env.RepoRoot)
	return s
}
