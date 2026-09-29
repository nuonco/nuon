package signal

import (
	"fmt"
	"runtime/debug"
)

type SignalErrInit struct {
	Err error
}

func (e *SignalErrInit) Error() string {
	return fmt.Sprintf("signal init failed: %v", e.Err)
}

func (e *SignalErrInit) Unwrap() error {
	return e.Err
}

type SignalErrValidate struct {
	Err error
}

func (e *SignalErrValidate) Error() string {
	return fmt.Sprintf("signal validate failed: %v", e.Err)
}

func (e *SignalErrValidate) Unwrap() error {
	return e.Err
}

type SignalErrExecute struct {
	Err error
}

func (e *SignalErrExecute) Error() string {
	return fmt.Sprintf("signal execute failed: %v", e.Err)
}

func (e *SignalErrExecute) Unwrap() error {
	return e.Err
}

type SignalErrPanic struct {
	Value      any
	Phase      string
	StackTrace string
}

func NewSignalErrPanic(value any, phase string) *SignalErrPanic {
	return &SignalErrPanic{
		Value:      value,
		Phase:      phase,
		StackTrace: string(debug.Stack()),
	}
}

func (e *SignalErrPanic) Error() string {
	return fmt.Sprintf("signal panicked during %s: %v\n%s", e.Phase, e.Value, e.StackTrace)
}
