package waypoint

import (
	"context"
	"fmt"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

func (r *repo) GetRunner(ctx context.Context, runnerID string) (*waypointv1.Runner, error) {
	client, err := r.WaypointClientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get waypoint client: %w", err)
	}

	resp, err := r.getRunner(ctx, client, runnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	return resp, nil
}

func (r *repo) getRunner(ctx context.Context, client waypointClient, runnerID string) (*waypointv1.Runner, error) {
	return client.GetRunner(ctx, &waypointv1.GetRunnerRequest{
		RunnerId: runnerID,
	})
}
