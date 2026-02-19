package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetComponentBuildForPolicyEvalRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetComponentBuildForPolicyEval(ctx context.Context, req GetComponentBuildForPolicyEvalRequest) (*app.ComponentBuild, error) {
	return a.getComponentBuildForPolicyEval(ctx, req.ID)
}

func (a *Activities) getComponentBuildForPolicyEval(ctx context.Context, buildID string) (*app.ComponentBuild, error) {
	var bld app.ComponentBuild
	res := a.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("ComponentConfigConnection").
		Preload("ComponentConfigConnection.Component").
		Preload("ComponentConfigConnection.Component.App").
		Preload("ComponentConfigConnection.Component.App.Org").
		Preload("ComponentConfigConnection.AppConfig").
		Preload("ComponentConfigConnection.AppConfig.PoliciesConfig").
		Preload("ComponentConfigConnection.AppConfig.PoliciesConfig.Policies", func(db *gorm.DB) *gorm.DB {
			return db.Limit(100)
		}).
		First(&bld, "id = ?", buildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component build: %w", res.Error)
	}

	return &bld, nil
}
