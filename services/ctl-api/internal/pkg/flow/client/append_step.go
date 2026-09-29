package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type AppendStepRequest struct {
	WorkflowID     string
	Name           string
	Signal         qsignal.Signal
	ExecutionType  app.WorkflowStepExecutionType
	StepTargetType string
	StepTargetID   string
	Retryable      bool
	Skippable      bool
	SkipOnFailure  bool
}

type AppendStepResponse struct {
	WorkflowID string
	GroupID    string
	StepID     string
}

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
		SkipOnFailure:  req.SkipOnFailure,
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
