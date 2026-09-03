package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type CreateAppConfigInput struct {
	AppID                  string        `json:"app_id" validate:"required"`
	OrgID                  string        `json:"org_id" validate:"required"`
	AppBranchID            string        `json:"app_branch_id" validate:"required"`
	CreatedByID            string        `json:"created_by_id" validate:"required"`
	IntermediateConfigJSON string        `json:"intermediate_config_json" validate:"required"`
	SourceConfigJSON       string        `json:"source_config_json,omitempty"`
	Labels                 labels.Labels `json:"labels,omitempty"`
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

	pendingStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusPending))
	pendingStatus.StatusHumanDescription = "pending sync"

	appConfig := &app.AppConfig{
		AppID:              req.AppID,
		OrgID:              req.OrgID,
		CreatedByID:        req.CreatedByID,
		AppBranchID:        generics.NewNullString(req.AppBranchID),
		Status:             app.AppConfigStatusPending,
		StatusDescription:  "pending sync",
		StatusV2:           pendingStatus,
		IntermediateConfig: &blobstore.Blob{},
		SourceConfig:       &blobstore.Blob{},
	}
	if req.Labels != nil {
		appConfig.Labels = req.Labels
	}
	appConfig.IntermediateConfig.Set(req.IntermediateConfigJSON)
	if req.SourceConfigJSON != "" {
		appConfig.SourceConfig.Set(req.SourceConfigJSON)
		appConfig.SourceConfig.SetContentType("application/json")
	}

	dbCtx := blobstore.WithBlobService(ctx, a.blobSvc)
	if res := a.db.WithContext(dbCtx).Create(appConfig); res.Error != nil {
		return nil, fmt.Errorf("unable to create app config: %w", res.Error)
	}

	return &CreateAppConfigOutput{
		AppConfigID: appConfig.ID,
	}, nil
}
