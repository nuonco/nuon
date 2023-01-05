package services

import (
	"fmt"

	"github.com/google/uuid"
)

type InvalidIDErr struct {
	value string
	err   error
}

func (i InvalidIDErr) Unwrap() error {
	return i.err
}

func (i InvalidIDErr) Error() string {
	return fmt.Sprintf("%s is not a valid uuid", i.value)
}

// parseID: parse the provided value as a uuid
func parseID(val string) (uuid.UUID, error) {
	uid, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, InvalidIDErr{
			value: val,
			err:   err,
		}
	}

	return uid, nil
}
