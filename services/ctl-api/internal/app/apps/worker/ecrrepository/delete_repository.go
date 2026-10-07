package ecrrepository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/generics"
)

type DeleteRepositoryRequest struct {
	OrgID string `validate:"required" json:"org_id"`
	AppID string `validate:"required" json:"app_id"`
}

func (r DeleteRepositoryRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type DeleteRepositoryResponse struct{}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 1m
func (a *Activities) DeleteRepository(ctx context.Context, req *DeleteRepositoryRequest) (*DeleteRepositoryResponse, error) {
	if err := req.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	awsCfg, err := credentials.Fetch(ctx, &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     a.cfg.ManagementIAMRoleARN,
			SessionName: "ctl-api-app-management",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch credentials: %w", err)
	}

	ecrClient := ecr.NewFromConfig(awsCfg)
	_, err = ecrClient.DeleteRepository(ctx, &ecr.DeleteRepositoryInput{
		RepositoryName: generics.ToPtr(req.OrgID + "/" + req.AppID),
		Force:          true,
	})
	if err != nil && !isRepositoryNotFoundException(err) {
		return nil, fmt.Errorf("failed to delete ecr repo: %w", err)
	}

	return &DeleteRepositoryResponse{}, nil
}
