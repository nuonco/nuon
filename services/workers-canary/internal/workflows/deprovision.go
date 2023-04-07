package workflows

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"go.temporal.io/sdk/workflow"
)

type deprovisionStep struct {
	name string
	fn   func(workflow.Context, string, *canaryv1.DeprovisionRequest) (*canaryv1.Step, error)
}

func (w *wkflow) Deprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.DeprovisionResponse, error) {
	l := workflow.GetLogger(ctx)

	canaryID := shortid.New()
	resp := &canaryv1.DeprovisionResponse{
		Steps:    make([]*canaryv1.Step, 0),
		CanaryId: canaryID,
	}
	if err := req.Validate(); err != nil {
		return resp, err
	}

	steps := []deprovisionStep{
		{
			"org",
			w.deprovisionOrg,
		},
		{
			"app",
			w.deprovisionApp,
		},
		//{
		//"install",
		//w.deprovisionInstall,
		//},
		//{
		//"docker-pull-deployment",
		//w.deprovisionDeployment,
		//},
	}

	for _, step := range steps {
		stepResp, err := step.fn(ctx, canaryID, req)
		if err != nil {
			return resp, fmt.Errorf("unable to deprovision %s %w", step.name, err)
		}

		resp.Steps = append(resp.Steps, stepResp)
		l.Info("successfully executed %s step", step.name)
	}

	return resp, nil
}

func (w *wkflow) deprovisionOrg(ctx workflow.Context, canaryID string, canaryReq *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
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

func (w *wkflow) deprovisionApp(ctx workflow.Context, canaryID string, req *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): implement deprovision app workflow
	return nil, fmt.Errorf("not implemented")
}

func (w *wkflow) deprovisionInstall(ctx workflow.Context, canaryID string, req *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): build out actual install request
	return nil, fmt.Errorf("not implemented")
	l := workflow.GetLogger(ctx)
	wkflowReq := &installsv1.DeprovisionRequest{
		OrgId:     canaryID,
		AppId:     canaryID,
		InstallId: canaryID,
	}

	l.Info("provisioning app", "request", wkflowReq)
	workflowID, err := w.startWorkflow(ctx, "apps", "Deprovision", wkflowReq)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}

	pollResp, err := w.pollWorkflow(ctx, "apps", "Deprovision", workflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	}
	l.Info("successfully got app response", "response", pollResp)
	return nil, nil
}

func (w *wkflow) deprovisionDeployment(ctx workflow.Context, canaryID string, req *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): build out actual deployment request
	return nil, fmt.Errorf("not implemented")
}
