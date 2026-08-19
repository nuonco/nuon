package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type AdoptQueuedComponentBuildRequest struct {
	ComponentID    string `validate:"required"`
	AppConfigID    string `validate:"required"`
	AppBranchRunID string
}

type AdoptQueuedComponentBuildOutput struct {
	BuildID string `json:"build_id"`
}

// AdoptQueuedComponentBuild returns the queued build pre-created by the app
// config syncer for this component on the given app config, stamping it with
// the branch run ID. Empty BuildID when no queued build exists.
//
// @temporal-gen-v2 activity
func (a *Activities) AdoptQueuedComponentBuild(ctx context.Context, req AdoptQueuedComponentBuildRequest) (*AdoptQueuedComponentBuildOutput, error) {
	out := &AdoptQueuedComponentBuildOutput{}

	if err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ccc app.ComponentConfigConnection
		err := tx.
			Select("id").
			Where(app.ComponentConfigConnection{
				AppConfigID: req.AppConfigID,
				ComponentID: req.ComponentID,
			}).
			Order("created_at DESC").
			First(&ccc).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("unable to get config connection: %w", err)
		}

		var bld app.ComponentBuild
		err = tx.
			Select("id").
			Where(app.ComponentBuild{
				ComponentConfigConnectionID: ccc.ID,
				Status:                      app.ComponentBuildStatusQueued,
			}).
			Order("created_at DESC").
			First(&bld).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("unable to get queued build: %w", err)
		}

		if req.AppBranchRunID != "" {
			if res := tx.
				Model(&app.ComponentBuild{}).
				Where("id = ?", bld.ID).
				Update("app_branch_run_id", req.AppBranchRunID); res.Error != nil {
				return fmt.Errorf("unable to link build to branch run: %w", res.Error)
			}
		}

		out.BuildID = bld.ID
		return nil
	}); err != nil {
		return nil, err
	}

	return out, nil
}
