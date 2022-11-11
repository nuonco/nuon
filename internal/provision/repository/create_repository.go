package repository

import (
	"context"

	"github.com/go-playground/validator/v10"
)

type CreateRepositoryRequest struct{}

func (r CreateRepositoryRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateRepositoryResponse struct{}

func (a *Activities) CreateRepository(ctx context.Context, req CreateRepositoryRequest) (CreateRepositoryResponse, error) {
	var resp CreateRepositoryResponse
	return resp, nil
}
