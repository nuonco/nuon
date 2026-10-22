package client

import (
	"context"
	"strings"
	"time"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

func (c *Client) ResumeRunnerJobSignal(ctx context.Context, signalID string) (app.Status, error) {
	var qs app.QueueSignal
	for {
		if err := c.db.WithContext(ctx).Where(app.QueueSignal{ID: signalID}).First(&qs).Error; err != nil {
			return "", err
		}
		switch qs.Status.Status {
		case app.StatusSuccess, app.StatusError, app.StatusCancelled:
			return qs.Status.Status, nil
		}
		handle, err := handler.UpdateWithStart(ctx, c.tClient, &qs, handler.UpdateWithStartOptions{
			UpdateName: handler.ResumeRunnerJobUpdateName, WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
		var status app.Status
		if err == nil {
			err = handle.Get(ctx, &status)
		}
		if err == nil {
			return status, nil
		}
		if !strings.Contains(err.Error(), "handler already finished") && !strings.Contains(err.Error(), "handler is executing") && !strings.Contains(err.Error(), "unknown update") && !strings.Contains(err.Error(), "closing workflow") {
			return "", err
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}
