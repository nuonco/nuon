package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

type CreateInstallStackVersionRequest struct {
	InstallID      string `validate:"required"`
	InstallStackID string `validate:"required"`
	AppConfigID    string `validate:"required"`
	Region         string `json:"region"`
	StackName      string `json:"stack_name"`
	Platform       string `json:"platform"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateInstallStackVersion(ctx context.Context, req *CreateInstallStackVersionRequest) (*app.InstallStackVersion, error) {
	return a.helpers.CreateInstallStackVersion(ctx, &helpers.CreateInstallStackVersionRequest{
		InstallID:      req.InstallID,
		InstallStackID: req.InstallStackID,
		AppConfigID:    req.AppConfigID,
		Region:         req.Region,
		StackName:      req.StackName,
		Platform:       req.Platform,
	})
}
