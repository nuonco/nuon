package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type GetLatestComponentBuildRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ID
func (a *Activities) GetLatestComponentBuild(ctx context.Context, req GetLatestComponentBuildRequest) (*app.ComponentBuild, error) {
	builds, err := a.componentsHelpers.GetComponentLatestBuilds(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		// Terminal: retrying will not produce a build. Step-level auto-retry
		// short-circuits on this marker (see signal.IsTerminalError).
		return nil, signal.NewTerminalError(
			"no_component_build",
			"No active build found for component (id %s). Ensure there is an active build for the component before retrying.",
			req.ID,
		)
	}

	// We only asked for one ID, so we should only have one build
	return &builds[0], nil
}
