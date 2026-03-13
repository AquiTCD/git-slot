package tui

import (
	"io"
	"os"

	"golang.org/x/term"
)

func IsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func IsNoColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}
