package workflows

import (
	"context"
	"fmt"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	tclient "go.temporal.io/sdk/client"
)

func (wfClient *workflowsClient) ScheduleCanaryProvision(ctx context.Context, id, schedule string, req *canaryv1.ProvisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		ID:           id,
		CronSchedule: schedule,
		TaskQueue:    DefaultTaskQueue,
		Memo: map[string]interface{}{
			"canary-id":  req.CanaryId,
			"started-by": "nuonctl",
		},
	}

	_, err := wfClient.TemporalClient.ExecuteWorkflowInNamespace(ctx, "canary", opts, "Provision", req)
	if err != nil {
		return fmt.Errorf("unable to provision canary: %w", err)
	}

	return nil
}

func (wfClient *workflowsClient) UnscheduleCanaryProvision(ctx context.Context, id string) error {
	if err := wfClient.TemporalClient.CancelWorkflowInNamespace(ctx, "canary", id, ""); err != nil {
		return fmt.Errorf("unable to provision canary: %w", err)
	}

	return nil
}

func (wfClient *workflowsClient) TriggerCanaryProvision(ctx context.Context, req *canaryv1.ProvisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
		Memo: map[string]interface{}{
			"canary-id":  req.CanaryId,
			"started-by": "nuonctl",
		},
	}

	_, err := wfClient.TemporalClient.ExecuteWorkflowInNamespace(ctx, "canary", opts, "Provision", req)
	if err != nil {
		return fmt.Errorf("unable to provision canary: %w", err)
	}

	return nil
}

func (wfClient *workflowsClient) TriggerCanaryDeprovision(ctx context.Context, req *canaryv1.DeprovisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
		Memo: map[string]interface{}{
			"canary-id":  req.CanaryId,
			"started-by": "nuonctl",
		},
	}

	_, err := wfClient.TemporalClient.ExecuteWorkflowInNamespace(ctx, "canary", opts, "Deprovision", req)
	if err != nil {
		return fmt.Errorf("unable to provision canary: %w", err)
	}

	return nil
}
