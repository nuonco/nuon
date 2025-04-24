package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type UpdateRunnerGroupSettings struct {
	RunnerID           string `json:"runner_id"`
	LocalAWSIAMRoleARN string `json:"runner_iam_role_arn"`
}

// @temporal-gen activity
func (a *Activities) UpdateRunnerGroupSettings(ctx context.Context, req *UpdateRunnerGroupSettings) error {
	return nil
	runner, err := a.getRunner(ctx, req.RunnerID)
	if err != nil {
		return err
	}

	groupSettings := app.RunnerGroupSettings{
		ID: runner.RunnerGroup.Settings.ID,
	}
	res := a.db.WithContext(ctx).
		Model(&groupSettings).
		Updates(app.RunnerGroupSettings{
			LocalAWSIAMRoleARN: req.LocalAWSIAMRoleARN,
		})
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update settings")
	}

	if res.RowsAffected < 1 {
		return generics.TemporalGormError(gorm.ErrRecordNotFound, "unable to find settings")
	}

	return nil
}
