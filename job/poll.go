package job

import (
	"context"
	"fmt"
)

// Poll polls a job to completion or error , writing events to the provided event writer func Poll(ctx context.Context,
func Poll(ctx context.Context, client waypointClientJobPoller, jobID string, writer EventWriter) error {
	poller := waypointDeploymentJobPollerImpl{}

	streamClient, err := poller.getWaypointDeploymentJobStream(ctx, client, jobID)
	if err != nil {
		return fmt.Errorf("unable to get job stream: %w", err)
	}

	if err := poller.consumeWaypointDeploymentJobStream(ctx, jobID, streamClient, writer); err != nil {
		return err
	}

	return nil
}
