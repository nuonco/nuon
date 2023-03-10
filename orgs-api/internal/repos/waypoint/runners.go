package waypoint

import (
	"context"
	"fmt"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

func (r *repo) ListRunners(ctx context.Context) (*waypointv1.ListRunnersResponse, error) {
	client, err := r.WaypointClientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get waypoint client: %w", err)
	}

	resp, err := r.listRunners(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("unable to list runners: %w", err)
	}
	return resp, nil
}

func (r *repo) listRunners(ctx context.Context, client waypointClient) (*waypointv1.ListRunnersResponse, error) {
	return client.ListRunners(ctx, &waypointv1.ListRunnersRequest{})
}
