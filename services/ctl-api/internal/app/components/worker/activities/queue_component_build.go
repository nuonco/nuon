package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

type QueueComponentBuildRequest struct {
	ComponentID string `validate:"required"`
	OrgID       string `validate:"required"`
	CreatedByID string `validate:"required"`
}

func (a *Activities) QueueComponentBuild(ctx context.Context, req QueueComponentBuildRequest) (*app.ComponentBuild, error) {
	// set the orgID on the context, for all writes
	ctx = context.WithValue(ctx, "org_id", req.OrgID)
	ctx = context.WithValue(ctx, "account_id", req.CreatedByID)

	build, err := a.helpers.CreateComponentBuild(ctx, req.ComponentID, true, nil)
	if err != nil {
		return nil, fmt.Errorf("create component build: %w", err)
	}

	a.evClient.Send(ctx, req.ComponentID, &signals.Signal{
		Type:    signals.OperationBuild,
		BuildID: build.ID,
	})
	return build, nil
}
