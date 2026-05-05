package stderr

import (
	"errors"
	"fmt"
)

// StepDirective tells the workflow step framework what to do when an
// ErrUser of this kind reaches handleStepError. The empty value means
// "use the signal's normal auto-retry policy" — the historical default.
//
// v1 ships with Stop and Skip; the type is intentionally extensible so
// future directives (e.g. SkipDependents, RetryGroup) can be added without
// touching producer call sites.
type StepDirective string

const (
	// StepDirectiveDefault leaves the existing auto-retry policy in charge.
	// Producers that just want a normal failure should leave Directive unset.
	StepDirectiveDefault StepDirective = ""

	// StepDirectiveStop marks the step as errored without consuming the
	// auto-retry budget — retrying without external action will not change
	// the outcome (e.g., missing component build, plan superseded).
	StepDirectiveStop StepDirective = "stop"

	// StepDirectiveSkip marks the step as skipped (not errored) and
	// continues group execution — the failure represents "nothing to do
	// here" rather than a problem (e.g., teardown of an already-empty
	// component).
	StepDirectiveSkip StepDirective = "skip"
)

// NewStopError returns an ErrUser whose Directive instructs the step
// framework to stop without auto-retry. code must be lowercase snake_case.
// description is the user-facing copy rendered in the dashboard / CLI.
// fields and cause may be nil.
func NewStopError(code, description string, fields map[string]string, cause error) ErrUser {
	return newDirectiveError(StepDirectiveStop, code, description, fields, cause)
}

// NewSkipError returns an ErrUser whose Directive instructs the step
// framework to mark the step as skipped (StatusSkipped) and continue.
// Use this for "no work to do" outcomes (e.g., teardown of an already
// empty resource).
func NewSkipError(code, description string, fields map[string]string, cause error) ErrUser {
	return newDirectiveError(StepDirectiveSkip, code, description, fields, cause)
}

func newDirectiveError(d StepDirective, code, description string, fields map[string]string, cause error) ErrUser {
	if cause == nil {
		// Build a sensible underlying message from the code and description so
		// .Error() is meaningful even when the producer didn't supply a cause.
		msg := description
		if msg == "" {
			msg = string(d)
		}
		if code != "" {
			cause = fmt.Errorf("[%s] %s", code, msg)
		} else {
			cause = errors.New(msg)
		}
	}
	return ErrUser{
		Err:         cause,
		Description: description,
		Code:        code,
		Fields:      fields,
		Directive:   d,
	}
}

// IsUserError unwraps err looking for an ErrUser and returns it by value.
// Returns the zero value and false if err does not contain an ErrUser.
//
// This is a convenience over errors.As to avoid the address-of-zero-value
// boilerplate at every call site.
func IsUserError(err error) (ErrUser, bool) {
	var u ErrUser
	if errors.As(err, &u) {
		return u, true
	}
	return ErrUser{}, false
}

// DirectiveOf returns the StepDirective on the wrapped ErrUser, or
// StepDirectiveDefault when err does not carry an ErrUser.
func DirectiveOf(err error) StepDirective {
	if u, ok := IsUserError(err); ok {
		return u.Directive
	}
	return StepDirectiveDefault
}
