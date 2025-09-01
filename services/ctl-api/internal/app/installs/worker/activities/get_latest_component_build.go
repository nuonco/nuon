package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetLatestComponentBuildRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetLatestComponentBuild(ctx context.Context, req GetLatestComponentBuildRequest) (*app.ComponentBuild, error) {
	builds, err := a.componentsHelpers.GetComponentLatestBuilds(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		return nil, fmt.Errorf("no builds found for component ID %s", req.ID)
	}

	// We only asked for one ID, so we should only have one build
	return &builds[0], nil
}
