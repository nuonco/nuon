package iam

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type CreateDeploymentsBucketRoleRequest struct{}

func (r CreateDeploymentsBucketRoleRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateDeploymentsBucketRoleResponse struct{}

func (a *Activities) CreateDeploymentsBucketRole(
	ctx context.Context,
	req CreateDeploymentsBucketRoleRequest,
) (CreateDeploymentsBucketRoleResponse, error) {
	var resp CreateDeploymentsBucketRoleResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	return resp, nil
}
