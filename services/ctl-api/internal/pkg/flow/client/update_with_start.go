package client

import (
	"context"
	"strings"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

func (c *Client) updateWithStartUntilAccepted(
	ctx context.Context,
	qs *app.QueueSignal,
	updateName string,
	args ...any,
) (tclient.WorkflowUpdateHandle, error) {
	return updateWithStartUntilAccepted(ctx, updateName, func() (tclient.WorkflowUpdateHandle, error) {
		return handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
			UpdateName:   updateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args:         args,
		})
	})
}

func (c *Client) updateWithStartUntilCompleted(
	ctx context.Context,
	qs *app.QueueSignal,
	updateName string,
	result any,
	args ...any,
) error {
	var accepted bool
	return handler.RetryUpdate(ctx, updateName, func(err error) bool {
		return accepted && isColdHostUpdateRejection(err, updateName)
	}, func() error {
		handle, err := c.updateWithStartUntilAccepted(ctx, qs, updateName, args...)
		accepted = err == nil
		if err != nil {
			return err
		}
		return handle.Get(ctx, result)
	})
}

func updateWithStartUntilAccepted(
	ctx context.Context,
	updateName string,
	submit func() (tclient.WorkflowUpdateHandle, error),
) (tclient.WorkflowUpdateHandle, error) {
	var handle tclient.WorkflowUpdateHandle
	err := handler.RetryUpdate(ctx, updateName, func(err error) bool {
		return isColdHostUpdateStartRace(err, updateName)
	}, func() error {
		var err error
		handle, err = submit()
		return err
	})
	if err != nil {
		return nil, err
	}
	return handle, nil
}

func isColdHostUpdateStartRace(err error, updateName string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unknown update "+updateName+".") ||
		strings.Contains(msg, "aborted by closing workflow")
}

func isColdHostUpdateRejection(err error, updateName string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.HasPrefix(msg, "unknown update "+updateName+".") ||
		strings.HasPrefix(msg, "workflow update was aborted by closing workflow") ||
		strings.HasPrefix(msg, "update was aborted by closing workflow") ||
		strings.Contains(msg, executeflow.HostFinishedRejection)
}
