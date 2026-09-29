package ui

type ErrExitCode struct {
	Err  error
	Code string
	Exit int
}

func (e *ErrExitCode) Error() string { return e.Err.Error() }

func (e *ErrExitCode) Unwrap() error { return e.Err }

func (e *ErrExitCode) ExitCode() int { return e.Exit }
