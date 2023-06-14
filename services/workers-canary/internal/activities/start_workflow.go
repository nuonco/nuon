package activities

import (
	"context"
	"fmt"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1/activities/v1"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	tclient "go.temporal.io/sdk/client"
)

func (a *Activities) StartWorkflow(ctx context.Context, req *activitiesv1.StartWorkflowRequest) (*activitiesv1.StartWorkflowResponse, error) {
	obj, err := req.Request.UnmarshalNew()
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal any request: %w", err)
	}

	tClient, err := temporalclient.New(a.v,
		temporalclient.WithNamespace(req.Namespace),
		temporalclient.WithAddr(a.TemporalHost))
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	// trigger workflow
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflowsclient.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"started-by": "workers-canary",
		},
	}

	wkflow, err := tClient.ExecuteWorkflow(ctx, opts, req.WorkflowName, obj)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}
	return &activitiesv1.StartWorkflowResponse{
		WorkflowId: wkflow.GetID(),
	}, nil
}
