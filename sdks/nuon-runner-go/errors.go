package nuonrunner

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type statusErr struct {
	Status string
}

func (s statusErr) Error() string {
	return fmt.Sprintf("invalid status %s", s.Status)
}

type stderrResponse interface {
	error
	IsCode(int) bool
	IsServerError() bool
	GetPayload() *models.StderrErrResponse
}

// httpStatusErr is returned by hand-rolled HTTP helpers that do not go through
// the generated client, so IsUnauthorized / IsForbidden still work.
type httpStatusErr struct {
	code int
	body string
}

func (e *httpStatusErr) Error() string {
	return fmt.Sprintf("unexpected status %d: %s", e.code, e.body)
}

func (e *httpStatusErr) IsCode(code int) bool {
	return e.code == code
}

func (e *httpStatusErr) IsServerError() bool {
	return e.code >= http.StatusInternalServerError
}

func (e *httpStatusErr) GetPayload() *models.StderrErrResponse {
	return &models.StderrErrResponse{}
}

func statusError(code int, body string) error {
	return &httpStatusErr{code: code, body: body}
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
