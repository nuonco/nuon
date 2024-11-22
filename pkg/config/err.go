package config

import "errors"

type ErrConfig struct {
	Description string
	Err         error

	Warning bool
}

func (e ErrConfig) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Description
}

func IsWarningErr(err error) bool {
	var ec ErrConfig
	if errors.As(err, &ec) {
		return ec.Warning
	}

	return false
}
