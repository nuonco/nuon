package generics

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
)

func TemporalDoNotRetry(err error, args ...string) error {
	for _, arg := range args {
		err = errors.Wrap(err, arg)
	}

	return temporal.NewNonRetryableApplicationError("non retryable",
		fmt.Sprintf("%T", err),
		err)
}

func TemporalGormError(err error, args ...string) error {
	for _, arg := range args {
		err = errors.Wrap(err, arg)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return temporal.NewNonRetryableApplicationError("not found",
			fmt.Sprintf("%T", err),
			err)
	}

	return err
}

func IsGormErrRecordNotFound(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}

	// NOTE: if this was returned via a temporal activity, we want to check the string, as well.
	if strings.Contains(err.Error(), "not found") {
		return true
	}

	return false
}
