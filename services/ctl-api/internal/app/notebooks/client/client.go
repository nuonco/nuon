// Package client dispatches notebook cell runs to the warm, long-lived
// per-notebook Temporal workflow via update-with-start.
package client

import (
	"context"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"

	"github.com/pkg/errors"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/actions"
)

// notebookNamespace is the Temporal namespace the installs worker (which
// registers NotebookWorkflow) runs in.
const notebookNamespace = "installs"

type Client struct {
	tClient temporalclient.Client
}

type Params struct {
	fx.In

	TClient temporalclient.Client
}

func New(params Params) *Client {
	return &Client{tClient: params.TClient}
}

type RunCellRequest struct {
	NotebookID     string
	CellID         string
	IdempotencyKey string
	OrgID          string
	InstallID      string
	TriggeredByID  string
}

type RunCellResponse struct {
	NotebookCellRunID          string
	InstallActionWorkflowRunID string
}

// RunCell dispatches a cell run to the notebook's warm workflow, starting the
// workflow if it isn't already running. It blocks only until the "run-cell"
// update handler completes (the run is created + enqueued), not until the run
// finishes — so the HTTP caller returns fast.
func (c *Client) RunCell(ctx context.Context, req RunCellRequest) (*RunCellResponse, error) {
	startOp := c.tClient.NewWithStartWorkflowOperation(
		tclient.StartWorkflowOptions{
			ID:                       actions.NotebookWorkflowID(req.NotebookID),
			TaskQueue:                workflows.APITaskQueue,
			WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
			RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 0},
		},
		actions.NotebookWorkflowType,
		actions.NotebookWorkflowRequest{NotebookID: req.NotebookID},
	)

	handle, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, notebookNamespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   actions.NotebookWorkflowID(req.NotebookID),
			UpdateName:   actions.NotebookRunCellUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				actions.RunCellUpdate{
					CellID:         req.CellID,
					IdempotencyKey: req.IdempotencyKey,
					OrgID:          req.OrgID,
					InstallID:      req.InstallID,
					TriggeredByID:  req.TriggeredByID,
				},
			},
		},
		StartWorkflowOperation: startOp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to dispatch notebook cell run")
	}

	var result actions.RunCellUpdateResult
	if err := handle.Get(ctx, &result); err != nil {
		return nil, errors.Wrap(err, "error waiting for run-cell handler")
	}

	return &RunCellResponse{
		NotebookCellRunID:          result.NotebookCellRunID,
		InstallActionWorkflowRunID: result.InstallActionWorkflowRunID,
	}, nil
}
