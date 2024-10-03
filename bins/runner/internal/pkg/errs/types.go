package errs

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errbase"
	"github.com/nuonco/nuon-runner-go/models"
)

type ErrType string

const (
	ErrTypeHandler   ErrType = "handler"
	ErrTypeFramework ErrType = "framework"
)

// RunnerError is the root of the runner's error taxonomy. It is used when no more specific error type is applicable.
type RunnerError struct {
	cause error
}

func (e *RunnerError) Cause() error                            { return e.cause }
func (e *RunnerError) Unwrap() error                           { return e.cause }
func (e *RunnerError) SafeFormatError(p errbase.Printer) error { return e.cause }

func (e *RunnerError) Error() string {
	// TODO(sdboyer) this is the big question - do we inject more strings here, or not?
	return e.cause.Error()
}

// RunnerHandlerError is a wrapper that indicates the contained error chain was emitted from a runner job handler.
//
// Job handlers should NOT create this error directly - it is used by the runner framework itself to categorize errors. The type is exported only to enable sniffing
type RunnerHandlerError struct {
	RunnerError
	JobGroup models.AppRunnerJobGroup
	JobType  models.AppRunnerJobType
	JobStep  string
}

func WithHandlerError(err error, jobGroup models.AppRunnerJobGroup, jobStep string, jobType models.AppRunnerJobType) error {
	return &RunnerHandlerError{
		RunnerError: RunnerError{
			cause: err,
		},
		JobGroup: jobGroup,
		JobType:  jobType,
		JobStep:  jobStep,
	}
}

var _ error = (*RunnerHandlerError)(nil)
var _ fmt.Formatter = (*RunnerHandlerError)(nil)
var _ errbase.SafeFormatter = (*RunnerHandlerError)(nil)

func (e *RunnerHandlerError) Is(err error) bool {
	// TODO(sdboyer) this is not the way to use the Is system
	_, is := err.(*RunnerHandlerError)
	return is
}

func (e *RunnerHandlerError) Format(s fmt.State, verb rune) { errbase.FormatError(e, s, verb) }

func (e *RunnerHandlerError) ErrorTags() map[string]string {
	return map[string]string{
		"runner_err_area":  string(ErrTypeHandler),
		"runner_job_group": string(e.JobGroup),
		"runner_job_type":  string(e.JobType),
		"runner_job_step":  e.JobStep,
	}
}

func WithFrameworkError(err error, jobType string) error {
	return &RunnerFrameworkError{
		RunnerError: RunnerError{
			cause: errors.WithStackDepth(err, 1),
		},
		JobType: jobType,
	}
}

// RunnerFrameworkError is a wrapper that indicates the contained error chain was emitted from some part of the runner framework itself, not a handler.
//
// This error type is applied as a fallback during job processing if no other error type is applied.
//
// TODO(sdboyer) having an error type like this doesn't do much until we make our errors properly network-portable
type RunnerFrameworkError struct {
	RunnerError
	JobType string
}

var _ error = (*RunnerFrameworkError)(nil)
var _ fmt.Formatter = (*RunnerFrameworkError)(nil)
var _ errbase.SafeFormatter = (*RunnerFrameworkError)(nil)

func (e *RunnerFrameworkError) Is(err error) bool {
	// TODO(sdboyer) this is not the way to use the Is system
	_, is := err.(*RunnerFrameworkError)
	return is
}

func (e *RunnerFrameworkError) Format(s fmt.State, verb rune) { errbase.FormatError(e, s, verb) }

func (e *RunnerFrameworkError) ErrorTags() map[string]string {
	ret := map[string]string{
		"runner_err_area": string(ErrTypeFramework),
	}
	if e.JobType != "" {
		ret["runner_job_type"] = e.JobType
	}
	return ret
}
