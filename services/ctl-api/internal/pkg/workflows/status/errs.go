package status

import (
	"fmt"
)

type StatusUpdateErr struct {
	origErr   error
	statusErr error
}

func (s *StatusUpdateErr) Error() string {
	if s.origErr == nil {
		return fmt.Sprintf("unable to update status: %s", s.statusErr.Error())
	}

	return fmt.Sprintf("%s\nunable to update status: %s", s.origErr, s.statusErr)
}

func (s *StatusUpdateErr) Unwrap() error {
	if s.origErr == nil {
		return s.statusErr
	}
	return s.origErr
}

// When you are updating a status, you often are setting an
func WrapStatusErr(origErr, statusErr error) error {
	return &StatusUpdateErr{
		origErr:   origErr,
		statusErr: statusErr,
	}
}

func StatusErr(statusErr error) error {
	return &StatusUpdateErr{
		statusErr: statusErr,
	}
}
