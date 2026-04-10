package cmd

import (
	"fmt"

	"github.com/AquiTCD/git-slot/internal/config"
)

// parseExecArgv parses arguments after the git-slot binary name.
// Expected forms: exec -- <command>...  or  exec <slot> -- <command>...
func parseExecArgv(argv []string, cfg *config.Config) (slotExplicit bool, slotName string, cmdArgs []string, err error) {
	if cfg == nil {
		return false, "", nil, fmt.Errorf("internal error: nil config")
	}
	execIdx := -1
	for i := 1; i < len(argv); i++ {
		if argv[i] == "exec" {
			execIdx = i
			break
		}
	}
	if execIdx < 0 {
		return false, "", nil, fmt.Errorf("exec subcommand not found in argv")
	}
	seg := argv[execIdx+1:]
	dash := -1
	for i, a := range seg {
		if a == "--" {
			dash = i
			break
		}
	}
	if dash < 0 {
		return false, "", nil, fmt.Errorf("git slot exec: use '--' before the command (e.g. git slot exec -- npm test)")
	}
	head := seg[:dash]
	cmdArgs = seg[dash+1:]
	if len(cmdArgs) == 0 {
		return false, "", nil, fmt.Errorf("git slot exec: no command after '--'")
	}
	switch len(head) {
	case 0:
		return false, "", cmdArgs, nil
	case 1:
		if cfg.FindSlot(head[0]) == nil {
			return false, "", nil, fmt.Errorf("unknown slot %q", head[0])
		}
		return true, head[0], cmdArgs, nil
	default:
		return false, "", nil, fmt.Errorf("git slot exec: at most one slot name is allowed before '--'")
	}
}
