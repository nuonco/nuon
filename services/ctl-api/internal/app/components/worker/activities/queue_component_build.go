package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type QueueComponentBuildRequest struct {
	ComponentID string `validate:"required"`
}

func (a *Activities) QueueComponentBuild(ctx context.Context, req GetComponentAppRequest) (*app.ComponentBuild, error) {
	build, err := a.helpers.CreateComponentBuild(ctx, req.ComponentID, true, nil)
	if err != nil {
		return nil, fmt.Errorf("create component build: %w", err)
	}

	a.hooks.BuildCreated(ctx, req.ComponentID, build.ID)
	return build, nil
}
