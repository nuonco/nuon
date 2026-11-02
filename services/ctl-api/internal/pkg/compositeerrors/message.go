package compositeerrors

import (
	"errors"

	"go.temporal.io/sdk/temporal"
)

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if multiple, ok := err.(interface{ Unwrap() []error }); ok {
		for _, cause := range multiple.Unwrap() {
			if cause != nil {
				return ErrorMessage(cause)
			}
		}
	}
	if cause := errors.Unwrap(err); cause != nil {
		return ErrorMessage(cause)
	}
	var applicationErr *temporal.ApplicationError
	if errors.As(err, &applicationErr) {
		return RedactDiagnosticSecrets(applicationErr.Message())
	}
	return RedactDiagnosticSecrets(err.Error())
}
