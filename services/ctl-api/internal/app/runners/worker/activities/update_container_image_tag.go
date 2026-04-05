package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateContainerImageTagRequest struct {
	RunnerID string `validate:"required"`
	Tag      string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
func (a *Activities) UpdateContainerImageTag(ctx context.Context, req UpdateContainerImageTagRequest) error {
	var runner app.Runner
	if res := a.db.WithContext(ctx).
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		First(&runner, "id = ?", req.RunnerID); res.Error != nil {
		return fmt.Errorf("unable to get runner: %w", res.Error)
	}

	if runner.RunnerGroup.Settings.ID == "" {
		return fmt.Errorf("runner %s has no settings", req.RunnerID)
	}

	if res := a.db.WithContext(ctx).
		Model(&app.RunnerGroupSettings{}).
		Where("id = ?", runner.RunnerGroup.Settings.ID).
		Update("container_image_tag", req.Tag); res.Error != nil {
		return fmt.Errorf("unable to update container image tag: %w", res.Error)
	}

	return nil
}
