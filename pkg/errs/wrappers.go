package errs

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errbase"
	"github.com/cockroachdb/errors/withstack"
)

// WithUserFacing wraps another error with a message intended for direct consumption by an end user.
//
// Use this function when there is exactly one known user who will consume the message. For example:
//
//   - DO use this in a CLI command where it's known that a CLI user will see the message
//   - DO NOT use this on the server side of ctl-api, where the caller may be CLI or terraform
//
// This function will also ensure a stack trace is collected. Stack traces are emitted to sentry,
// but never shown to end users.
//
// If there is no underlying error to be wrapped, use [NewUserFacing] instead.
func WithUserFacing(err error, format string, args ...any) error {
	// TODO (sdboyer) is it worth creating a stack trace only if there isn't already one in the tree?
	// return errors.WithHint(errors.WithStackDepth(err, 1), msg)
	return errors.WithStackDepth(errors.WithHint(err, fmt.Sprintf(format, args...)), 1)
}

// NewUserFacing creates a new error with a message intended for direct consumption by an end user.
//
// This is the equivalent of [WithUserFacing], but for use when there's no underlying error to be wrapped.
func NewUserFacing(format string, args ...any) error {
	// Duplicating the error string as the hint string is silly, but we do it for consistency
	return errors.WithHint(errors.NewWithDepthf(1, format, args...), fmt.Sprintf(format, args...))
}

// HasNuonStackTrace reports whether the provided error contains at least one stack trace that originated in Nuon code.
func HasNuonStackTrace(err error) bool {
	var stacks []*withstack.ReportableStackTrace
	visitAllMulti(err, func(c error) {
		st := withstack.GetReportableStackTrace(c)
		if st != nil {
			stacks = append(stacks, st)
		}
	})

	for _, st := range stacks {
		for _, fr := range st.Frames {
			if strings.Contains(fr.Module, "powertoolsdev") || strings.Contains(fr.Module, "nuonco") {
				return true
			}
		}
	}
	return false
}

func visitAllMulti(err error, f func(error)) {
	f(err)
	if e := errbase.UnwrapOnce(err); e != nil {
		visitAllMulti(e, f)
	}
	for _, e := range errbase.UnwrapMulti(err) {
		visitAllMulti(e, f)
	}
}
