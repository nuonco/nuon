package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
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

	var res executeflow.AppendStepResponse
	err = c.updateWithStartUntilCompleted(ctx, qs, "append-step", &res, executeflow.AppendStepRequest{
		Name:           req.Name,
		Signal:         signaldb.SignalData{Signal: req.Signal},
		ExecutionType:  req.ExecutionType,
		StepTargetType: req.StepTargetType,
		StepTargetID:   req.StepTargetID,
		Retryable:      req.Retryable,
		Skippable:      req.Skippable,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send append-step update: %w", err)
	}

	return &AppendStepResponse{
		WorkflowID: res.WorkflowID,
		GroupID:    res.GroupID,
		StepID:     res.StepID,
	}, nil
}
