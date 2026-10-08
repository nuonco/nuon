package ecrrepository

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/azure/acr"
)

type DeleteACRRepositoryRequest struct {
	OrgID string `validate:"required" json:"org_id"`
	AppID string `validate:"required" json:"app_id"`
}

func (r DeleteACRRepositoryRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type DeleteACRRepositoryResponse struct{}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 1m
func (a *Activities) DeleteACRRepository(ctx context.Context, req *DeleteACRRepositoryRequest) (*DeleteACRRepositoryResponse, error) {
	if err := req.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	err := acr.DeleteRepository(ctx, nil, a.cfg.ManagementACRRegistryURL, req.OrgID+"/"+req.AppID, zap.L())
	if err != nil {
		return nil, fmt.Errorf("failed to delete acr repository: %w", err)
	}

	return &DeleteACRRepositoryResponse{}, nil
}
