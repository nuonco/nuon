package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type QueueComponentBuildRequest struct {
	ComponentID string `validate:"required"`
	OrgID       string `validate:"required"`
}

func (a *Activities) QueueComponentBuild(ctx context.Context, req QueueComponentBuildRequest) (*app.ComponentBuild, error) {
	// set the orgID on the context, for all writes
	ctx = context.WithValue(ctx, "org_id", req.OrgID)

	build, err := a.helpers.CreateComponentBuild(ctx, req.ComponentID, true, nil)
	if err != nil {
		return nil, fmt.Errorf("create component build: %w", err)
	}

	a.hooks.BuildCreated(ctx, req.ComponentID, build.ID)
	return build, nil
}
