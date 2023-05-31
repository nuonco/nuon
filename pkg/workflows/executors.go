package workflows

import (
	"context"
	"fmt"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	tclient "go.temporal.io/sdk/client"
)

func (wfClient *workflowsClient) ExecCreatePlan(ctx context.Context, req *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: ExecutorsTaskQueue,
		Memo: map[string]interface{}{
			"started-by": "nuonctl",
		},
	}

	resp := &planv1.CreatePlanResponse{}
	fut, err := wfClient.TemporalClient.ExecuteWorkflowInNamespace(ctx, "orgs", opts, "CreatePlan", req)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	if err := fut.Get(ctx, resp); err != nil {
		return nil, fmt.Errorf("unable to get plan: %w", err)
	}

	return resp, nil
}
