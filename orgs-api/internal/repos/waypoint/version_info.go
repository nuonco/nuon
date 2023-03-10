package waypoint

import (
	"context"
	"fmt"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (r *repo) GetVersionInfo(ctx context.Context) (*waypointv1.GetVersionInfoResponse, error) {
	client, err := r.WaypointClientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get waypoint client: %w", err)
	}

	resp, err := r.getVersionInfo(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("unable to get version info: %w", err)
	}

	return resp, nil
}

func (r *repo) getVersionInfo(ctx context.Context, client waypointClient) (*waypointv1.GetVersionInfoResponse, error) {
	return client.GetVersionInfo(ctx, &emptypb.Empty{})
}
