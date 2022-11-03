package job

import (
	"context"
	"fmt"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type waypointDeploymentJobPoller interface {
	getWaypointDeploymentJobStream(context.Context, waypointClientJobPoller, string) (gen.Waypoint_GetJobStreamClient, error)
	consumeWaypointDeploymentJobStream(context.Context, string, waypointClientJobStreamReceiver, EventWriter) error
}

var _ waypointDeploymentJobPoller = (*waypointDeploymentJobPollerImpl)(nil)

type waypointDeploymentJobPollerImpl struct{}

func (w *waypointDeploymentJobPollerImpl) getWaypointDeploymentJobStream(
	ctx context.Context,
	client waypointClientJobPoller,
	jobID string,
) (gen.Waypoint_GetJobStreamClient, error) {
	streamClient, err := client.GetJobStream(ctx, &gen.GetJobStreamRequest{
		JobId: jobID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get job stream client: %w", err)
	}

	return streamClient, nil
}

func (w *waypointDeploymentJobPollerImpl) consumeWaypointDeploymentJobStream(
	ctx context.Context,
	jobID string,
	client waypointClientJobStreamReceiver,
	evWriter EventWriter,
) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		var resp *gen.GetJobStreamResponse
		resp, err := client.Recv()
		if err != nil {
			return fmt.Errorf("error while receiving response: %w", err)
		}

		// handle err
		ev := waypointJobStreamResponseToWaypointEvent(jobID, resp)
		if err := evWriter.Write(ev); err != nil {
			return err
		}

		wpErr := waypointJobEventToErr(ev)
		switch wpErr {
		case errWaypointJobEventNoop:
			continue
		case nil:
			return nil
		default:
			return wpErr
		}
	}
}

type waypointClientJobPoller interface {
	GetJobStream(context.Context, *gen.GetJobStreamRequest, ...grpc.CallOption) (gen.Waypoint_GetJobStreamClient, error)
}

type waypointClientJobStreamReceiver interface {
	Recv() (*gen.GetJobStreamResponse, error)
}
