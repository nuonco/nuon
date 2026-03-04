package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type CreateAppConfigInput struct {
	AppID                  string `json:"app_id" validate:"required"`
	OrgID                  string `json:"org_id" validate:"required"`
	AppBranchID            string `json:"app_branch_id" validate:"required"`
	CreatedByID            string `json:"created_by_id" validate:"required"`
	IntermediateConfigJSON string `json:"intermediate_config_json" validate:"required"`
}

type CreateAppConfigOutput struct {
	AppConfigID string `json:"app_config_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
func (a *Activities) createAppConfig(ctx context.Context, req *CreateAppConfigInput) (*CreateAppConfigOutput, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	appConfig := &app.AppConfig{
		AppID:              req.AppID,
		OrgID:              req.OrgID,
		CreatedByID:        req.CreatedByID,
		AppBranchID:        generics.NewNullString(req.AppBranchID),
		Status:             app.AppConfigStatusPending,
		StatusDescription:  "pending sync",
		IntermediateConfig: &blobstore.Blob{},
	}
	appConfig.IntermediateConfig.Set(req.IntermediateConfigJSON)

	if res := a.db.WithContext(ctx).Create(appConfig); res.Error != nil {
		return nil, fmt.Errorf("unable to create app config: %w", res.Error)
	}

	return &CreateAppConfigOutput{
		AppConfigID: appConfig.ID,
	}, nil
}
