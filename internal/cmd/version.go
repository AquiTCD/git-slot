package cmd

import (
	"fmt"
	"io"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func printVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "git-slot version %s (commit: %s, built: %s)\n", version, commit, date)
}
