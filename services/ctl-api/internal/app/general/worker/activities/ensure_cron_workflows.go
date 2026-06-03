package activities

import (
	"context"
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/workflows"
)

type EnsureCronWorkflowsRequest struct{}

type EnsureCronWorkflowsResponse struct {
	SweepStarted   bool `json:"sweep_started"`
	MetricsStarted bool `json:"metrics_started"`
	CleanupStarted bool `json:"cleanup_started"`
}

// EnsureCronWorkflows starts (or replaces) the enqueuer-sweep and
// general-metrics-cron workflows as top-level Temporal cron workflows.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) EnsureCronWorkflows(ctx context.Context, _ EnsureCronWorkflowsRequest) (*EnsureCronWorkflowsResponse, error) {
	resp := &EnsureCronWorkflowsResponse{}

	// Start (or replace) the enqueuer sweep cron.
	sweepOpts := tclient.StartWorkflowOptions{
		ID:                    "enqueuer-sweep",
		TaskQueue:             workflows.APITaskQueue,
		CronSchedule:          "* * * * *",
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	type sweepReq struct{}
	if _, err := a.tClient.ExecuteWorkflowInNamespace(ctx, "general", sweepOpts, "EnqueuerSweep", sweepReq{}); err != nil {
		return nil, fmt.Errorf("unable to start enqueuer sweep workflow: %w", err)
	}
	resp.SweepStarted = true
	a.logger.Info("enqueuer sweep cron started/replaced", zap.String("workflow-id", "enqueuer-sweep"))

	// Start (or replace) the general metrics cron.
	metricsOpts := tclient.StartWorkflowOptions{
		ID:                    "general-metrics-cron",
		TaskQueue:             workflows.APITaskQueue,
		CronSchedule:          "*/1 * * * *",
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	if _, err := a.tClient.ExecuteWorkflowInNamespace(ctx, "general", metricsOpts, "Metrics"); err != nil {
		return nil, fmt.Errorf("unable to start general metrics workflow: %w", err)
	}
	resp.MetricsStarted = true
	a.logger.Info("general metrics cron started/replaced", zap.String("workflow-id", "general-metrics-cron"))

	// Start (or replace) the daily queue-signal cleanup cron.
	cleanupOpts := tclient.StartWorkflowOptions{
		ID:                    "general-queue-signal-cleanup-cron",
		TaskQueue:             workflows.APITaskQueue,
		CronSchedule:          "0 0 * * *",
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	if _, err := a.tClient.ExecuteWorkflowInNamespace(ctx, "general", cleanupOpts, "CleanupQueueSignals"); err != nil {
		return nil, fmt.Errorf("unable to start queue signal cleanup workflow: %w", err)
	}
	resp.CleanupStarted = true
	a.logger.Info("queue signal cleanup cron started/replaced", zap.String("workflow-id", "general-queue-signal-cleanup-cron"))

	return resp, nil
}
