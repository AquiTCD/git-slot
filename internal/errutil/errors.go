package errutil

// ExitError is an interface for errors that should cause the program to exit
// with a specific exit code.
type ExitError interface {
	error
	ExitCode() int
}

type basicExitError struct {
	msg  string
	code int
}

func (e *basicExitError) Error() string { return e.msg }
func (e *basicExitError) ExitCode() int { return e.code }

// NewExitError creates a new error that returns the given exit code.
func NewExitError(msg string, code int) ExitError {
	return &basicExitError{msg: msg, code: code}
}
