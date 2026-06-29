package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateInstallStackVersionRunRequest struct {
	RunID     string                        `json:"run_id" validate:"required"`
	RunType   app.StackVersionRunType       `json:"run_type"`
	RoleDiff  *app.StackVersionRunRoleDiff  `json:"role_diff,omitempty"`
	InputDiff *app.StackVersionRunInputDiff `json:"input_diff,omitempty"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateInstallStackVersionRun(ctx context.Context, req UpdateInstallStackVersionRunRequest) error {
	updates := map[string]interface{}{
		"run_type":   req.RunType,
		"role_diff":  req.RoleDiff,
		"input_diff": req.InputDiff,
	}
	if res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersionRun{}).
		Where("id = ?", req.RunID).
		Updates(updates); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update install stack version run")
	}
	return nil
}
