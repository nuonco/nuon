package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	approvalplan "github.com/nuonco/nuon/pkg/plans/types/approval_plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CheckNoopPlanRequest struct {
	StepTargetID string           `validate:"required"`
	WorkflowType app.WorkflowType `json:",omitempty"`
	InstallID    string           `json:",omitempty"`
	ComponentID  string           `json:",omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 5
func (a *Activities) CheckNoopPlan(ctx context.Context, req *CheckNoopPlanRequest) (bool, error) {
	if sandbox, err := a.isSandboxModeJob(ctx, req.StepTargetID); err == nil && sandbox {
		return false, nil
	}

	plan, err := a.getApprovalPlan(ctx, req.StepTargetID)
	if err != nil {
		a.recordDriftPlanEvaluationForTarget(ctx, req.WorkflowType, req.InstallID, req.ComponentID, false, "load_plan", err)
		return false, errors.Wrap(err, "unable to get approval plan")
	}

	isNoop, err := plan.IsNoopPlan()
	a.recordDriftPlanEvaluationForTarget(ctx, req.WorkflowType, req.InstallID, req.ComponentID, isNoop, "evaluate_plan", err)
	return isNoop, err
}

func (a *Activities) isSandboxModeJob(ctx context.Context, stepTargetID string) (bool, error) {
	runnerJob, err := a.getRunnerJob(ctx, &GetRunnerJobRequest{RunnerJobOwnerID: stepTargetID})
	if err != nil {
		return false, err
	}

	var runner app.Runner
	res := a.db.WithContext(ctx).
		Preload("RunnerGroup.Settings").
		Where(app.Runner{ID: runnerJob.RunnerID}).
		First(&runner)
	if res.Error != nil {
		return false, res.Error
	}

	return runner.RunnerGroup.Settings.SandboxMode, nil
}

func (a *Activities) getApprovalPlan(ctx context.Context, stepTargetID string) (*ApprovalPlan, error) {
	runnerJob, err := a.getRunnerJob(ctx, &GetRunnerJobRequest{RunnerJobOwnerID: stepTargetID})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to fetch runner job, step target id: %s", stepTargetID))
	}

	runnerJobExecution, err := a.getRunnerJobExecution(ctx, GetRunnerJobExecutionRequest{
		RunnerJobID: runnerJob.ID,
	})
	if err != nil {
		return nil, err
	}

	runnerJobExecutionResult, err := a.getRunnerJobExecutionResult(ctx, GetRunnerJobExecutionResultRequest{
		RunnerJobExecutionID: runnerJobExecution.ID,
	})
	if err != nil {
		return nil, err
	}

	// we're only using content display currently since we're only dealing with terraform and sandbox plans
	decompressedContentDisplay, err := a.decompressRunnerJobExecutionResult(runnerJobExecutionResult.ContentsDisplayGzip)
	if err != nil {
		return nil, err
	}

	plan := ApprovalPlan{
		RunnerJobType:        runnerJob.Type,
		RunnerJobID:          runnerJob.ID,
		RunnerJobExecutionID: runnerJobExecution.ID,
		PlanContents:         decompressedContentDisplay,
	}

	return &plan, nil
}

type ApprovalPlan struct {
	RunnerJobType        app.RunnerJobType `json:"runner_job_type" temporaljson:"runner_job_type,omitempty"`
	RunnerJobID          string            `json:"runner_job_id" temporaljson:"runner_job_id,omitempty"`
	RunnerJobExecutionID string            `json:"runner_job_execution_id" temporaljson:"runner_job_execution_id,omitempty"`
	PlanContents         []byte            `json:"plan_contents" temporaljson:"plan_contents,omitempty"`
}

func (p *ApprovalPlan) IsNoopPlan() (bool, error) {
	switch p.RunnerJobType {
	case app.RunnerJobTypeSandboxTerraform, app.RunnerJobTypeSandboxTerraformPlan:
		plan := approvalplan.NewSandboxRunApprovalPlan(p.PlanContents)
		return plan.IsNoop()
	case app.RunnerJobTypeSandboxPulumi:
		plan := approvalplan.NewPulumiApprovalPlan(p.PlanContents)
		return plan.IsNoop()
	case app.RunnerJobTypeTerraformDeploy:
		plan := approvalplan.NewTerraformApprovalPlan(p.PlanContents)
		return plan.IsNoop()
	case app.RunnerJobTypeKubrenetesManifestDeploy:
		plan := approvalplan.NewKubernetesApprovalPlan(p.PlanContents)
		return plan.IsNoop()
	case app.RunnerJobTypeHelmChartDeploy:
		plan := approvalplan.NewHelmApprovalPlen(p.PlanContents)
		return plan.IsNoop()
	case app.RunnerJobTypePulumiDeploy:
		plan := approvalplan.NewPulumiApprovalPlan(p.PlanContents)
		return plan.IsNoop()
	default:
		return false, fmt.Errorf("unsupported approval plan request, runner job type %s", p.RunnerJobType)
	}
}
