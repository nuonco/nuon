package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

const (
	updateWithStartMaxAttempts  = 15
	updateWithStartInitialDelay = 50 * time.Millisecond
	updateWithStartMaxDelay     = 5 * time.Second
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
	delay := updateWithStartInitialDelay
	for attempt := 1; attempt <= updateWithStartMaxAttempts; attempt++ {
		handle, err := c.updateWithStartUntilAccepted(ctx, qs, updateName, args...)
		if err != nil {
			return err
		}
		if err := handle.Get(ctx, result); err == nil {
			return nil
		} else if !isColdHostUpdateRejection(err, updateName) {
			return err
		}
		if attempt == updateWithStartMaxAttempts {
			return fmt.Errorf("%s handler not registered after %d attempts", updateName, updateWithStartMaxAttempts)
		}
		if err := waitForUpdateRetry(ctx, delay); err != nil {
			return err
		}
		delay *= 2
		if delay > updateWithStartMaxDelay {
			delay = updateWithStartMaxDelay
		}
	}
	return nil
}

func updateWithStartUntilAccepted(
	ctx context.Context,
	updateName string,
	submit func() (tclient.WorkflowUpdateHandle, error),
) (tclient.WorkflowUpdateHandle, error) {
	delay := updateWithStartInitialDelay
	var lastErr error
	for attempt := 1; attempt <= updateWithStartMaxAttempts; attempt++ {
		handle, err := submit()
		if err == nil {
			return handle, nil
		}
		if !isColdHostUpdateStartRace(err, updateName) {
			return nil, err
		}
		lastErr = err
		if attempt == updateWithStartMaxAttempts {
			break
		}

		if err := waitForUpdateRetry(ctx, delay); err != nil {
			return nil, err
		}
		delay *= 2
		if delay > updateWithStartMaxDelay {
			delay = updateWithStartMaxDelay
		}
	}

	return nil, fmt.Errorf("%s handler not registered after %d attempts: %w", updateName, updateWithStartMaxAttempts, lastErr)
}

func waitForUpdateRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
		strings.HasPrefix(msg, "update was aborted by closing workflow")
}
