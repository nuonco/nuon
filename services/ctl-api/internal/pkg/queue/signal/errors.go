package signal

import "fmt"

// SignalErrInit wraps an error that occurred during signal initialization.
type SignalErrInit struct {
	Err error
}

func (e *SignalErrInit) Error() string {
	return fmt.Sprintf("signal init failed: %v", e.Err)
}

func (e *SignalErrInit) Unwrap() error {
	return e.Err
}

// SignalErrValidate wraps an error that occurred during signal validation.
type SignalErrValidate struct {
	Err error
}

func (e *SignalErrValidate) Error() string {
	return fmt.Sprintf("signal validate failed: %v", e.Err)
}

func (e *SignalErrValidate) Unwrap() error {
	return e.Err
}

// SignalErrExecute wraps an error that occurred during signal execution.
type SignalErrExecute struct {
	Err error
}

func (e *SignalErrExecute) Error() string {
	return fmt.Sprintf("signal execute failed: %v", e.Err)
}

func (e *SignalErrExecute) Unwrap() error {
	return e.Err
}

// SignalErrBlocked indicates a phase was blocked by a lifecycle before-phase hook.
type SignalErrBlocked struct {
	Phase  SignalPhase
	Reason string
}

func (e *SignalErrBlocked) Error() string {
	return fmt.Sprintf("blocked by lifecycle hook during %s: %s", e.Phase, e.Reason)
}

// SignalErrPanic wraps a panic that occurred during signal processing.
type SignalErrPanic struct {
	Value any
	Phase string // "init", "validate", or "execute"
}

func (e *SignalErrPanic) Error() string {
	return fmt.Sprintf("signal panicked during %s: %v", e.Phase, e.Value)
}
