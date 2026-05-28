package nuon

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type stderrResponse interface {
	error
	IsCode(int) bool
	IsServerError() bool
	GetPayload() *models.StderrErrResponse
}

// httpAPIError is a simple error type for hand-written SDK methods that
// satisfies the stderrResponse interface so IsNotFound/IsForbidden/etc work.
type httpAPIError struct {
	statusCode int
	body       string
}

func (e *httpAPIError) Error() string {
	return fmt.Sprintf("unexpected status code %d: %s", e.statusCode, e.body)
}

func (e *httpAPIError) IsCode(code int) bool {
	return e.statusCode == code
}

func (e *httpAPIError) IsServerError() bool {
	return e.statusCode >= 500
}

func (e *httpAPIError) GetPayload() *models.StderrErrResponse {
	return &models.StderrErrResponse{
		Error:       e.body,
		Description: fmt.Sprintf("HTTP %d", e.statusCode),
	}
}

func newHTTPAPIError(statusCode int, body string) error {
	return &httpAPIError{statusCode: statusCode, body: body}
}

// ToUserError returns the error as a user error if possible
func ToUserError(inputErr error) (*models.StderrErrResponse, bool) {
	var (
		stderr stderrResponse
		ok     bool
	)
	for {
		stderr, ok = inputErr.(stderrResponse)
		if ok {
			break
		}

		inputErr = errors.Unwrap(inputErr)
		if inputErr == nil {
			return nil, false
		}
	}

	payload := stderr.GetPayload()
	if !payload.UserError {
		return nil, false
	}

	return payload, true
}

func IsUnauthorized(err error) bool {
	stderr, ok := err.(stderrResponse)
	if !ok {
		return false
	}

	return stderr.IsCode(401)
}

func IsForbidden(err error) bool {
	stderr, ok := err.(stderrResponse)
	if !ok {
		return false
	}

	return stderr.IsCode(403)
}

func IsNotFound(err error) bool {
	stderr, ok := err.(stderrResponse)
	if !ok {
		return false
	}

	return stderr.IsCode(404)
}

func IsBadRequest(err error) bool {
	stderr, ok := err.(stderrResponse)
	if !ok {
		return false
	}

	return stderr.IsCode(400)
}

func IsServerError(err error) bool {
	stderr, ok := err.(stderrResponse)
	if !ok {
		return false
	}

	return stderr.IsServerError()
}

// ToAPIError extracts a user-friendly error message from any API error response.
// Unlike ToUserError, this returns a message for all API errors, not just user errors.
// Returns the description if available, otherwise the error field, otherwise empty string and false.
func ToAPIError(inputErr error) (string, bool) {
	var (
		stderr stderrResponse
		ok     bool
	)
	for {
		stderr, ok = inputErr.(stderrResponse)
		if ok {
			break
		}

		inputErr = errors.Unwrap(inputErr)
		if inputErr == nil {
			return "", false
		}
	}

	payload := stderr.GetPayload()
	if payload == nil {
		return "", false
	}

	if payload.Description != "" {
		return payload.Description, true
	}
	if payload.Error != "" {
		return payload.Error, true
	}

	return "", false
}
