package retry

import "errors"

type nonRetryableErr struct {
	err error
}

func (e nonRetryableErr) Error() string {
	return e.err.Error()
}

func (e nonRetryableErr) Unwrap() error {
	return e.err
}

func AsNonRetryable(err error) error {
	return nonRetryableErr{
		err: err,
	}
}

func IsNonRetryable(err error) bool {
	var nre nonRetryableErr
	return errors.As(err, &nre)
}
