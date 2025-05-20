package orgiam

import (
	"errors"

	"github.com/aws/smithy-go"

	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func isEntityExistsException(err error) bool {
	if err == nil {
		return false
	}

	entityExistsErr := &iam_types.EntityAlreadyExistsException{}
	return errors.As(err, &entityExistsErr)
}

func isLimitExceededError(err error) bool {
	if err == nil {
		return false
	}

	limitExceededErr := &iam_types.LimitExceededException{}
	return errors.As(err, &limitExceededErr)
}

const (
	noEntityFoundCode string = "NoSuchEntity"
)

func isNotFoundErr(err error) bool {
	var ae smithy.APIError
	if errors.As(err, &ae) {
		return ae.ErrorCode() == noEntityFoundCode
	}

	return false
}
