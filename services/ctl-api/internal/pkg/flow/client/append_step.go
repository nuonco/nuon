package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// AppendStepRequest is the input for appending a step to a resident workflow.
type AppendStepRequest struct {
	// WorkflowID is the resident execute-workflow to append to.
	WorkflowID string
	// Name is the human-readable step/group name (e.g. the cell label).
	Name string
	// Signal is the work to run for the appended step (e.g. an
	// actionworkflowrun signal for a notebook cell).
	Signal qsignal.Signal
	// ExecutionType / StepTargetType / Retryable / Skippable are the step
	// metadata the caller computes the same way installSignalStep does, so the
	// low-level flow package stays free of the installs signal taxonomy.
	ExecutionType  app.WorkflowStepExecutionType
	StepTargetType string
	Retryable      bool
	Skippable      bool
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

	h, err := handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "append-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			executeflow.AppendStepRequest{
				Name:           req.Name,
				Signal:         signaldb.SignalData{Signal: req.Signal},
				ExecutionType:  req.ExecutionType,
				StepTargetType: req.StepTargetType,
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
