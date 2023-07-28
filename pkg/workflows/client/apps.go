package client

import (
	"context"
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	tclient "go.temporal.io/sdk/client"
)

func (w *workflowsClient) TriggerAppProvision(ctx context.Context, req *appsv1.ProvisionRequest) (string, error) {
	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-provision", req.AppId),
		TaskQueue: DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"started-by": w.Agent,
		},
	}

	wfRun, err := w.TemporalClient.ExecuteWorkflowInNamespace(ctx, "apps", opts, "Provision", req)
	if err != nil {
		return "", err
	}

	return wfRun.GetID(), err
}
