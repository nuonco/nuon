package activities

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	approvalplan "github.com/powertoolsdev/mono/pkg/plans/types/approval_plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CheckNoopPlanRequest struct {
	StepTargetID string `validate:"required"`
}

// @temporal-gen activity
// @max-retries 1
func (a *Activities) CheckNoopPlan(ctx context.Context, req *CheckNoopPlanRequest) (bool, error) {
	plan, err := a.getApprovalPlan(ctx, req.StepTargetID)
	if err != nil {
		return false, errors.Wrap(err, "unable to get approval plan")
	}

	return plan.IsNoopPlan()
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
	decompressedContentDisplay, err := a.decompressRunnerJobExecutionResult(ctx, runnerJobExecutionResult.ContentsDisplayGzip)
	if err != nil {
		return nil, err
	}

	plan := ApprovalPlan{
		RunnerJobType: runnerJob.Type,
		PlanContents:  decompressedContentDisplay,
	}

	return &plan, nil
}

type ApprovalPlan struct {
	RunnerJobType app.RunnerJobType `json:"runner_job_type" temporaljson:"runner_job_type,omitempty"`
	PlanContents  []byte            `json:"plan_contents" temporaljson:"plan_contents,omitempty"`
}

func (p *ApprovalPlan) IsNoopPlan() (bool, error) {
	switch p.RunnerJobType {
	case app.RunnerJobTypeSandboxTerraform, app.RunnerJobTypeSandboxTerraformPlan:
		plan := approvalplan.NewSandboxRunApprovalPlan(p.PlanContents)
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
	default:
		return false, fmt.Errorf("unsupported approval plan request, runner job type %s", p.RunnerJobType)
	}
}
