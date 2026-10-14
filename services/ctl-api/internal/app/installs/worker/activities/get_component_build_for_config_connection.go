package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
)

type GetComponentBuildForConfigConnectionRequest struct {
	ComponentConfigConnectionID string `validate:"required"`
}

// GetComponentBuildForConfigConnection resolves the build for a specific
// config connection. Returns (nil, nil) when none exists yet.
//
// @temporal-gen-v2 activity
// @by-field ComponentConfigConnectionID
func (a *Activities) GetComponentBuildForConfigConnection(ctx context.Context, req GetComponentBuildForConfigConnectionRequest) (*app.ComponentBuild, error) {
	return buildselection.Component(ctx, a.db, req.ComponentConfigConnectionID)
}
