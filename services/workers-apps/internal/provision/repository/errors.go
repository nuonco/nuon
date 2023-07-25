package repository

import (
	"errors"

	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

func isEntityExistsException(err error) bool {
	if err == nil {
		return false
	}

	entityExistsErr := &ecr_types.RepositoryAlreadyExistsException{}
	if errors.As(err, &entityExistsErr) {
		return true
	}

	return true
}
