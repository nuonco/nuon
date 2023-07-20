package iam

import (
	"errors"

	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func isEntityExistsException(err error) bool {
	if err == nil {
		return false
	}

	entityExistsErr := &iam_types.EntityAlreadyExistsException{}
	if errors.As(err, &entityExistsErr) {
		return true
	}

	return true
}
