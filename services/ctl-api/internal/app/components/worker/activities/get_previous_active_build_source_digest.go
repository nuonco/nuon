package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetPreviousActiveBuildSourceDigestRequest struct {
	// ComponentID is the component whose history we search.
	ComponentID string `validate:"required"`
	// ExcludeBuildID is the current build (which may itself already exist in
	// the DB at planning time). It is excluded from the lookup so we always
	// find a strictly prior build.
	ExcludeBuildID string `validate:"required"`
}

type GetPreviousActiveBuildSourceDigestResponse struct {
	// SourceDigest is the source manifest digest of the most recent prior
	// Active ComponentBuild for the component. Empty when no prior active
	// build exists, or when the prior build has no SourceDigest recorded.
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
