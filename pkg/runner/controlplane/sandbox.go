package controlplane

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/runner/jobs"
)

// isSandboxableBuildHandler reports whether a control-plane job handler
// publishes an artifact to the plan's destination registry. Sandbox-mode orgs
// never provision real registry infrastructure, so plan.Dst carries fake
// values and any real push/auth against it fails. noop-build and
// fetch-image-metadata are excluded: the former never pushes to Dst, and the
// latter's consumer expects a compressed metadata result rather than generic
// sandbox outputs.
func isSandboxableBuildHandler(handler jobs.JobHandler) bool {
	switch handler.Name() {
	case "container-image-build",
		"helm-build",
		"terraform-build",
		"pulumi-build",
		"kubernetes-manifest-build",
		"sandbox-build":
		return true
	default:
		return false
	}
}

func (e *Executor) getSandboxMode(ctx context.Context, jobID string) (*plantypes.SandboxMode, error) {
	planJSON, err := e.client.GetJobPlanJSON(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("unable to get job plan: %w", err)
	}

	var plan plantypes.MinSandboxMode
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil, fmt.Errorf("unable to parse sandbox mode from job plan: %w", err)
	}
	if plan.SandboxMode != nil && plan.SandboxMode.Enabled {
		return plan.SandboxMode, nil
	}

	var compositePlan plantypes.CompositePlan
	if err := json.Unmarshal([]byte(planJSON), &compositePlan); err != nil {
		return nil, fmt.Errorf("unable to parse composite job plan: %w", err)
	}
	inner := compositePlan.Inner()
	if inner == nil {
		return nil, nil
	}
	innerJSON, err := json.Marshal(inner)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal inner job plan: %w", err)
	}
	if err := json.Unmarshal(innerJSON, &plan); err != nil {
		return nil, fmt.Errorf("unable to parse sandbox mode from inner job plan: %w", err)
	}

	if plan.SandboxMode == nil || !plan.SandboxMode.Enabled {
		return nil, nil
	}
	return plan.SandboxMode, nil
}

func (e *Executor) executeSandboxBuild(
	ctx context.Context,
	job *models.AppRunnerJob,
	execution *models.AppRunnerJobExecution,
	sandboxMode *plantypes.SandboxMode,
) error {
	outputs := sandboxMode.Outputs
	if outputs == nil {
		outputs = map[string]any{}
	}

	updated, err := e.client.UpdateJobExecution(ctx, job.ID, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusInDashProgress,
	})
	if err != nil {
		return fmt.Errorf("unable to mark sandbox build in progress: %w", err)
	}
	if err := terminalExecutionError(updated.Status); err != nil {
		return err
	}

	if _, err := e.client.CreateJobExecutionOutputs(ctx, job.ID, execution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{
		Outputs: outputs,
	}); err != nil {
		return fmt.Errorf("unable to write sandbox build outputs: %w", err)
	}

	if _, err := e.client.CreateJobExecutionResult(ctx, job.ID, execution.ID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}); err != nil {
		return fmt.Errorf("unable to write sandbox build result: %w", err)
	}

	updated, err = e.client.UpdateJobExecution(ctx, job.ID, execution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return fmt.Errorf("unable to mark sandbox build finished: %w", err)
	}
	return terminalExecutionError(updated.Status)
}
