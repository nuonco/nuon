package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppBranchRunBuildsCompletedInput struct {
	RunID           string `json:"run_id" validate:"required"`
	BuildsCompleted bool   `json:"builds_completed"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) UpdateAppBranchRunBuildsCompleted(ctx context.Context, input *UpdateAppBranchRunBuildsCompletedInput) error {
	if err := a.v.Struct(input); err != nil {
		return errors.Wrap(err, "invalid request")
	}

	var run app.AppBranchRun
	if err := a.db.WithContext(ctx).First(&run, "id = ?", input.RunID).Error; err != nil {
		return errors.Wrap(err, "app branch run not found")
	}

	value := "false"
	if input.BuildsCompleted {
		value = "true"
	}

	if run.Labels == nil {
		run.Labels = labels.Labels{}
	}
	run.Labels.Merge(labels.Labels{
		app.AppBranchRunLabelBuildsCompleted: value,
	})

	if err := a.db.WithContext(ctx).
		Model(&run).
		Select("labels").
		Updates(&run).Error; err != nil {
		return fmt.Errorf("unable to update builds_completed label on run %s: %w", input.RunID, err)
	}

	return nil
}
