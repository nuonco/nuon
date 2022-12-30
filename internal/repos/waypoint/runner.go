package waypoint

import (
	"context"

	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

func (r *repo) GetRunner(ctx context.Context, runnerID string) (*waypointv1.Runner, error) {
	return nil, nil
}
