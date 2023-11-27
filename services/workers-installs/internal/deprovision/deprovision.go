package deprovision

import (
	"fmt"

	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/sandbox"
	"go.temporal.io/sdk/workflow"
)

func (w wkflow) createPlanRequest(runTyp planv1.SandboxInputType, req *installsv1.DeprovisionRequest) *planv1.CreatePlanRequest {
	return &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Sandbox{
			Sandbox: &planv1.SandboxInput{
				OrgId:           req.OrgId,
				AppId:           req.AppId,
				InstallId:       req.InstallId,
				RunId:           req.RunId,
				Type:            planv1.SandboxInputType_SANDBOX_INPUT_TYPE_DEPROVISION,
				AccountSettings: req.AccountSettings,
				SandboxSettings: req.SandboxSettings,
			},
		},
	}
}

func (w wkflow) deprovisionNoopBuild(ctx workflow.Context, req *installsv1.DeprovisionRequest) error {
	planReq := w.createPlanRequest(planv1.SandboxInputType_SANDBOX_INPUT_TYPE_NOOP_BUILD, req)
	planResp, err := sandbox.Plan(ctx, planReq)
	if err != nil {
		return fmt.Errorf("unable to create noop-build plan: %w", err)
	}

	_, err = sandbox.Execute(ctx, &executev1.ExecutePlanRequest{
		Plan: planResp.Plan,
	})
	if err != nil {
		return fmt.Errorf("unable to execute noop-build plan: %w", err)
	}

	return nil
}

func (w wkflow) deprovisionSandbox(ctx workflow.Context, req *installsv1.DeprovisionRequest) (*executev1.ExecutePlanResponse, error) {
	runTyp := planv1.SandboxInputType_SANDBOX_INPUT_TYPE_DEPROVISION

	planReq := w.createPlanRequest(runTyp, req)
	planResp, err := sandbox.Plan(ctx, planReq)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	execResp, err := sandbox.Execute(ctx, &executev1.ExecutePlanRequest{
		Plan: planResp.Plan,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute plan: %w", err)
	}

	return execResp, nil
}
