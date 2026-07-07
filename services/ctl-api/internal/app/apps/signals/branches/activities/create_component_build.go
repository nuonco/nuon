package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateComponentBuildRequest struct {
	BuildID     string `validate:"required"`
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateComponentBuild(ctx context.Context, req CreateComponentBuildRequest) (*app.ComponentBuild, error) {
	var existing app.ComponentBuild
	res := a.db.WithContext(ctx).Where(&app.ComponentBuild{ID: req.BuildID}).First(&existing)
	if res.Error == nil {
		return &existing, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to get component build: %w", res.Error)
	}

	cmp, err := a.componentHelpers.GetComponentByID(ctx, req.ComponentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	ctx = cctx.SetOrgIDContext(ctx, cmp.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, cmp.CreatedByID)

	build, err := a.componentHelpers.CreateComponentBuildWithID(ctx, req.BuildID, req.ComponentID, true, nil)
	if err != nil {
		res := a.db.WithContext(ctx).Where(&app.ComponentBuild{ID: req.BuildID}).First(&existing)
		if res.Error == nil {
			return &existing, nil
		}
		return nil, fmt.Errorf("unable to create component build: %w", err)
	}

	return build, nil
}
