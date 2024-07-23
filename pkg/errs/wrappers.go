package errs

import "github.com/cockroachdb/errors"

// UserFacingError wraps another error with a message intended for direct consumption by an end user.
//
// Use this function when there is exactly one known user who will consume the message. For example:
//
//   - DO use this in a CLI command where it's known that a CLI user will see the message
//   - DO NOT use this on the server side of ctl-api, where the caller may be CLI or terraform
//
// This function will also ensure a stack trace is collected. Stack traces are emitted to sentry,
// but never shown to end users.
func UserFacingError(err error, msg string) error {
	// TODO (sdboyer) is it worth creating a stack trace only if there isn't already one in the tree?
	// return errors.WithHint(errors.WithStackDepth(err, 1), msg)
	return errors.WithStackDepth(errors.WithHint(err, msg), 1)
}
