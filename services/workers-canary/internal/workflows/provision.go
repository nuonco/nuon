package workflows

import (
	"fmt"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) Provision(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.ProvisionResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("provisioning canary", "id", req.CanaryId)

	resp := &canaryv1.ProvisionResponse{
		Steps:    make([]*canaryv1.Step, 0),
		CanaryId: req.CanaryId,
	}
	if err := req.Validate(); err != nil {
		return resp, err
	}

	steps := []step{
		{
			"org",
			w.provisionOrg,
		},
		//{
		//"app",
		//w.provisionApp,
		//},
		//{
		//"install",
		//w.provisionInstall,
		//},
		//{
		//"docker-pull-deployment",
		//w.provisionDeployment,
		//},
	}

	for _, step := range steps {
		stepResp, err := step.fn(ctx, req.CanaryId, req)
		if err != nil {
			return resp, fmt.Errorf("unable to provision %s %w", step.name, err)
		}

		resp.Steps = append(resp.Steps, stepResp)
		l.Info("successfully executed %s step", step.name)
	}

	return resp, nil
}

func (w *wkflow) provisionOrg(ctx workflow.Context, canaryID string, canaryReq *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	l := workflow.GetLogger(ctx)
	wkflowReq := &orgsv1.SignupRequest{
		OrgId:  canaryID,
		Region: defaultRegion,
	}

	l.Info("provisioning org", "request", wkflowReq)
	workflowID, err := w.startWorkflow(ctx, "orgs", "Signup", wkflowReq)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}

	pollResp, err := w.pollWorkflow(ctx, "orgs", "Signup", workflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	}
	l.Info("successfully got org response", "response", pollResp)

	return pollResp.Step, nil
}

func (w *wkflow) provisionApp(ctx workflow.Context, canaryID string, req *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	l := workflow.GetLogger(ctx)
	wkflowReq := &appsv1.ProvisionRequest{
		OrgId: canaryID,
		AppId: canaryID,
	}

	l.Info("provisioning app", "request", wkflowReq)
	workflowID, err := w.startWorkflow(ctx, "apps", "Provision", wkflowReq)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}

	pollResp, err := w.pollWorkflow(ctx, "apps", "Provision", workflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	}
	l.Info("successfully got app response", "response", pollResp)
	return pollResp.Step, nil
}

func (w *wkflow) provisionInstall(ctx workflow.Context, canaryID string, req *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): build out actual install request
	return nil, fmt.Errorf("not implemented")
	l := workflow.GetLogger(ctx)
	wkflowReq := &installsv1.ProvisionRequest{
		OrgId:     canaryID,
		AppId:     canaryID,
		InstallId: canaryID,
	}

	l.Info("provisioning app", "request", wkflowReq)
	workflowID, err := w.startWorkflow(ctx, "apps", "Provision", wkflowReq)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}

	pollResp, err := w.pollWorkflow(ctx, "apps", "Provision", workflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	}
	l.Info("successfully got app response", "response", pollResp)
	return nil, nil
}

func (w *wkflow) provisionDeployment(ctx workflow.Context, canaryID string, req *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): build out actual deployment request
	return nil, fmt.Errorf("not implemented")
}
