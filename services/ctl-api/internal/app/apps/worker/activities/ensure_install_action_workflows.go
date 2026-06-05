package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/ensure"
)

type EnsureInstallActionWorkflowsRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppID
func (a *Activities) EnsureInstallActionWorkflows(ctx context.Context, req *EnsureInstallActionWorkflowsRequest) error {
	return ensure.ActionWorkflows(ctx, a.db, req.AppID, nil)
}
