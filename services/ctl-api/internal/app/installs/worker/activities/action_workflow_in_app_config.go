package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ActionWorkflowInAppConfigRequest struct {
	AppConfigID      string `validate:"required"`
	ActionWorkflowID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) ActionWorkflowInAppConfig(ctx context.Context, req ActionWorkflowInAppConfigRequest) (bool, error) {
	var count int64

	res := a.db.WithContext(ctx).
		Model(&app.ActionWorkflowConfig{}).
		Where(app.ActionWorkflowConfig{
			AppConfigID:      req.AppConfigID,
			ActionWorkflowID: req.ActionWorkflowID,
		}).
		Count(&count)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to check action workflow in app config")
	}

	return count > 0, nil
}
