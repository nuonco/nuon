package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	// A cold resident host registers its append-step update handler only after
	// the queue Handler workflow boots and runs RegisterUpdateHandlers. An
	// append issued in that window is rejected with an untyped "unknown update"
	// error; retry until the handler is live. Subsequent appends to a warm host
	// land on the first attempt.
	//
	// Retries use exponential backoff from appendStepInitialDelay, doubling each
	// attempt up to appendStepMaxDelay. This stays responsive for the common
	// warm/quick-boot case while spreading out the tail for a slow cold boot
	// (workflow closed -> fresh Handler boot -> RegisterUpdateHandlers). The
	// attempt count is sized so the cumulative wait is ~30s.
	appendStepMaxAttempts  = 15
	appendStepInitialDelay = 50 * time.Millisecond
	appendStepMaxDelay     = 5 * time.Second
)

// AppendStepRequest is the input for appending a step to a resident workflow.
type AppendStepRequest struct {
	// WorkflowID is the resident execute-workflow to append to.
	WorkflowID string
	// Name is the human-readable step/group name (e.g. the step label).
	Name string
	// Signal is the work to run for the appended step (e.g. an
	// actionworkflowrun signal for an appended step).
	Signal qsignal.Signal
	// ExecutionType / StepTargetType / Retryable / Skippable are the step
	// metadata the caller computes the same way installSignalStep does, so the
	// low-level flow package stays free of the installs signal taxonomy.
	ExecutionType  app.WorkflowStepExecutionType
	StepTargetType string
	// StepTargetID identifies the work the step drives (e.g. the action run id)
	// and doubles as the append idempotency key — a retry with the same id
	// returns the existing step instead of appending a duplicate.
	StepTargetID string
	Retryable    bool
	Skippable    bool
}

// AppendStepResponse reports the freshly-created group and step IDs.
type AppendStepResponse struct {
	WorkflowID string
	GroupID    string
	StepID     string
}

// AppendStep sends an "append-step" update to the resident execute-flow handler
// workflow for the given workflow. Uses update-with-start so the handler is
// (re)started if it has idled out, and waits for completion so the caller gets
// the created group/step IDs back.
func (c *Client) AppendStep(ctx context.Context, req *AppendStepRequest) (*AppendStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.WorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	var lastErr error
	delay := appendStepInitialDelay
	for attempt := 1; attempt <= appendStepMaxAttempts; attempt++ {
		resp, err := c.appendStepOnce(ctx, qs, req)
		if err == nil {
			return resp, nil
		}
		if !isColdHostAppendRace(err) {
			return nil, err
		}

		lastErr = err
		if attempt == appendStepMaxAttempts {
			break
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("append-step retry interrupted: %w", ctx.Err())
		case <-timer.C:
		}

		delay = nextAppendStepDelay(delay)
	}

	return nil, fmt.Errorf("append-step handler not registered after %d attempts: %w", appendStepMaxAttempts, lastErr)
}

// nextAppendStepDelay doubles the current cold-host retry delay, capping it at
// appendStepMaxDelay so the backoff tail stays bounded.
func nextAppendStepDelay(delay time.Duration) time.Duration {
	delay *= 2
	if delay > appendStepMaxDelay {
		return appendStepMaxDelay
	}
	return delay
}

// appendStepOnce issues a single update-with-start. The start operation uses
// USE_EXISTING so retries never spawn a duplicate host; append idempotency is
// keyed on StepTargetID inside the handler, so a retry that lands after the
// step already exists returns it rather than appending a duplicate.
func (c *Client) appendStepOnce(ctx context.Context, qs *app.QueueSignal, req *AppendStepRequest) (*AppendStepResponse, error) {
	h, err := handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "append-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			executeflow.AppendStepRequest{
				Name:           req.Name,
				Signal:         signaldb.SignalData{Signal: req.Signal},
				ExecutionType:  req.ExecutionType,
				StepTargetType: req.StepTargetType,
				StepTargetID:   req.StepTargetID,
				Retryable:      req.Retryable,
				Skippable:      req.Skippable,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send append-step update: %w", err)
	}

	var res executeflow.AppendStepResponse
	if err := h.Get(ctx, &res); err != nil {
		return nil, fmt.Errorf("error waiting for append-step handler: %w", err)
	}

	return &AppendStepResponse{
		WorkflowID: res.WorkflowID,
		GroupID:    res.GroupID,
		StepID:     res.StepID,
	}, nil
}

// isColdHostAppendRace reports whether err is a transient cold-start race
// against the resident host, safe to retry because USE_EXISTING re-attaches to
// the (re)warmed host and append idempotency on StepTargetID prevents
// duplicates. Two windows exist before the host is steady-state:
//
//   - "unknown update append-step": the queue Handler workflow booted but has
//     not yet run RegisterUpdateHandlers.
//   - "aborted by closing workflow": the append landed while the Handler was
//     continuing-as-new, closing the run the update attached to.
//
// Both are untyped/server strings (no typed error), so a narrow match is the
// only option. A steady-state warm host hits neither.
func isColdHostAppendRace(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "unknown update") && strings.Contains(msg, "append-step") {
		return true
	}
	return strings.Contains(msg, "aborted by closing workflow")
}
