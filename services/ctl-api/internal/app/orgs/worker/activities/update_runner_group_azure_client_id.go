package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateRunnerGroupAzureClientIDRequest struct {
	OrgID         string `validate:"required"`
	AzureClientID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
func (a *Activities) UpdateRunnerGroupAzureClientID(ctx context.Context, req UpdateRunnerGroupAzureClientIDRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.RunnerGroupSettings{}).
		Where("org_id = ?", req.OrgID).
		Update("org_azure_client_id", req.AzureClientID)
	if res.Error != nil {
		return fmt.Errorf("unable to update runner group azure client ID: %w", res.Error)
	}

	return nil
}
