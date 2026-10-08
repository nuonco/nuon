package ecrrepository

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

	return false
}

func isRepositoryNotFoundException(err error) bool {
	if err == nil {
		return false
	}

	notFoundErr := &ecr_types.RepositoryNotFoundException{}
	if errors.As(err, &notFoundErr) {
		return true
	}

	return false
}
