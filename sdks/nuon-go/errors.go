package nuon

import (
	"errors"
	"fmt"

	"github.com/nuonco/nuon-go/models"
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
