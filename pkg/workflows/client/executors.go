package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

func (w *workflowsClient) ExecCreatePlan(ctx context.Context, namespace string, req *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: ExecutorsTaskQueue,
		Memo: map[string]interface{}{
			"started-by": w.Agent,
		},
	}

	resp := &planv1.CreatePlanResponse{}
	fut, err := w.TemporalClient.ExecuteWorkflowInNamespace(ctx, namespace, opts, "CreatePlan", req)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	if err := fut.Get(ctx, resp); err != nil {
		return nil, fmt.Errorf("unable to get plan: %w", err)
	}

	return resp, nil
}
