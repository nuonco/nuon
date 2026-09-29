package queue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

func TestStopDuringQueueStartup(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	c, err := client.Dial(client.Options{
		HostPort:  os.Getenv("TEMPORAL_HOST_PORT"),
		Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
	})
	require.NoError(t, err)
	defer c.Close()

	id := "queue-stop-startup-" + uuid.NewString()
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	metadata := make(chan activities.UpdateQueueMetadataRequest, 1)
	w := worker.New(c, id, worker.Options{})
	w.RegisterWorkflowWithOptions(func(ctx workflow.Context) (bool, error) {
		q := &queue{queueID: id}
		finished, err := q.run(ctx)
		return finished && q.stopped, err
	}, workflow.RegisterOptions{Name: "StopDuringStartup"})
	w.RegisterActivityWithOptions(func(ctx context.Context, req activities.QueueExistsRequest) (bool, error) {
		started <- struct{}{}
		select {
		case <-release:
			return true, nil
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}, activity.RegisterOptions{Name: "QueueInternalQueueExists"})
	w.RegisterActivityWithOptions(func(ctx context.Context, req activities.UpdateQueueMetadataRequest) error {
		metadata <- req
		return nil
	}, activity.RegisterOptions{Name: "QueueInternalUpdateQueueMetadata"})
	require.NoError(t, w.Start())
	defer w.Stop()
	defer close(release)

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: id, TaskQueue: id}, "StopDuringStartup")
	require.NoError(t, err)
	defer func() {
		_ = c.TerminateWorkflow(context.Background(), id, run.GetRunID(), "test cleanup")
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("queue existence activity did not start")
	}
	update, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
		WorkflowID:   id,
		RunID:        run.GetRunID(),
		UpdateName:   StopUpdateName,
		Args:         []any{StopRequest{}},
		WaitForStage: client.WorkflowUpdateStageCompleted,
	})
	require.NoError(t, err)
	var response StopResponse
	require.NoError(t, update.Get(ctx, &response))
	release <- struct{}{}
	var stopped bool
	require.NoError(t, run.Get(ctx, &stopped))
	require.True(t, stopped)
	select {
	case req := <-metadata:
		require.Equal(t, id, req.QueueID)
		require.NotNil(t, req.Metadata["stopped_at"])
		_, err := time.Parse(time.RFC3339, *req.Metadata["stopped_at"])
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("stopped_at was not written")
	}
}
