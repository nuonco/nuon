package ui

// ErrExitCode wraps an error with a stable machine code for the agent
// envelope and a custom process exit code, honored by the command wrapper.
// It lets a command distinguish outcomes beyond the generic exit 1, e.g.
// "config synced but component builds failed".
type ErrExitCode struct {
	Err  error
	Code string
	Exit int
}

func (e *ErrExitCode) Error() string { return e.Err.Error() }

func (e *ErrExitCode) Unwrap() error { return e.Err }

func (e *ErrExitCode) ExitCode() int { return e.Exit }
