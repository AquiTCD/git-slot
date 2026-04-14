// Package gslenv holds the contract for the gsl shell wrapper's internal env marker.
package gslenv

import (
	"os"
	"strings"
)

// Name is the environment variable set to "1" by the gsl wrapper for command-substitution calls.
const Name = "GSL_FROM_WRAPPER"

// FromWrapper reports whether the current process was invoked from the gsl wrapper
// capture path (suppresses launch_shell so git-slot can print a path for cd).
func FromWrapper() bool {
	return os.Getenv(Name) == "1"
}

// StripFromEnv returns a copy of env without any entry whose key is Name.
func StripFromEnv(env []string) []string {
	prefix := Name + "="
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// StripFromOSEnviron is shorthand for StripFromEnv(os.Environ()).
func StripFromOSEnviron() []string {
	return StripFromEnv(os.Environ())
}
