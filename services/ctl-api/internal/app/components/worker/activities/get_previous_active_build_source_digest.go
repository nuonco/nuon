package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetPreviousActiveBuildSourceDigestRequest struct {
	ComponentID    string `validate:"required"`
	ExcludeBuildID string `validate:"required"`
}

type GetPreviousActiveBuildSourceDigestResponse struct {
	SourceDigest string
}

// @temporal-gen-v2 activity
// @by-field ComponentID
func (a *Activities) GetPreviousActiveBuildSourceDigest(ctx context.Context, req GetPreviousActiveBuildSourceDigestRequest) (*GetPreviousActiveBuildSourceDigestResponse, error) {
	var bld app.ComponentBuild

	res := a.db.WithContext(ctx).
		Joins("JOIN component_config_connections ON component_config_connections.id = component_builds.component_config_connection_id").
		Where("component_config_connections.component_id = ?", req.ComponentID).
		Where("component_builds.id <> ?", req.ExcludeBuildID).
		Where("component_builds.status = ?", app.ComponentBuildStatusActive).
		Where("component_builds.source_digest <> ''").
		Order("component_builds.created_at DESC").
		Limit(1).
		First(&bld)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return &GetPreviousActiveBuildSourceDigestResponse{}, nil
		}
		return nil, errors.Wrap(res.Error, "unable to get previous active component build")
	}

	return &GetPreviousActiveBuildSourceDigestResponse{
		SourceDigest: bld.SourceDigest,
	}, nil
}
