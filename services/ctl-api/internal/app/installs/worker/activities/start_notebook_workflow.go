package activities

import (
	"context"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/workflows"
)

// notebookWorkflowType and the "notebook-" ID prefix mirror the constants in
// the installs/worker/actions package. They are duplicated here (rather than
// imported) because actions already imports this activities package, and
// importing it back would create a cycle. Keep in sync with
// actions.NotebookWorkflowType / actions.NotebookWorkflowID.
const (
	notebookWorkflowType  = "NotebookWorkflow"
	notebookWorkflowIDPfx = "notebook-"
	notebookNamespace     = "installs"
)

type StartNotebookWorkflowRequest struct {
	NotebookID string `validate:"required"`
}

type StartNotebookWorkflowResponse struct {
	WorkflowID string
}

// startNotebookWorkflowRequest matches the JSON shape of
// actions.NotebookWorkflowRequest for a cold start (State omitted). Only the
// NotebookID is carried; the workflow initializes its own state.
type startNotebookWorkflowRequest struct {
	NotebookID string
}

// StartNotebookWorkflow starts the warm per-notebook workflow if it isn't
// already running, and reuses the existing run otherwise (USE_EXISTING). This
// is the queue-managed lifecycle entrypoint: a notebook-start signal calls it
// so the workflow is warm before the first cell run dispatches to it.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) StartNotebookWorkflow(ctx context.Context, req *StartNotebookWorkflowRequest) (*StartNotebookWorkflowResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, errors.Wrap(err, "invalid request")
	}

	workflowID := notebookWorkflowIDPfx + req.NotebookID
	opts := tclient.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                workflows.APITaskQueue,
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 0},
	}
	if _, err := a.tClient.ExecuteWorkflowInNamespace(ctx, notebookNamespace, opts,
		notebookWorkflowType, startNotebookWorkflowRequest{NotebookID: req.NotebookID}); err != nil {
		return nil, errors.Wrap(err, "unable to start notebook workflow")
	}

	return &StartNotebookWorkflowResponse{WorkflowID: workflowID}, nil
}
