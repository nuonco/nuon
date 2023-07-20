package workflows

import (
	"fmt"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"go.temporal.io/sdk/workflow"
)

type deprovisionStep struct {
	name string
	fn   func(workflow.Context, *canaryv1.DeprovisionRequest) (*canaryv1.Step, error)
}

func (w *wkflow) Deprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.DeprovisionResponse, error) {
	l := workflow.GetLogger(ctx)

	resp := &canaryv1.DeprovisionResponse{
		Steps:    make([]*canaryv1.Step, 0),
		CanaryId: req.CanaryId,
	}
	if err := req.Validate(); err != nil {
		return resp, err
	}

	w.sendNotification(ctx, notificationTypeDeprovisionStart, req.CanaryId, nil)
	steps := []deprovisionStep{
		{
			"install",
			w.deprovisionInstall,
		},
		{
			"org",
			w.deprovisionOrg,
		},
		//{
		//"docker-pull-deployment",
		//w.deprovisionDeployment,
		//},
	}

	for _, step := range steps {
		stepResp, err := step.fn(ctx, req)
		if err != nil {
			err = fmt.Errorf("unable to deprovision %s: %w", step.name, err)
			w.sendNotification(ctx, notificationTypeDeprovisionError, req.CanaryId, err)
			return resp, err
		}

		resp.Steps = append(resp.Steps, stepResp)
		l.Info("successfully executed %s step", step.name)
	}

	w.sendNotification(ctx, notificationTypeDeprovisionSuccess, req.CanaryId, nil)
	return resp, nil
}

func (w *wkflow) deprovisionOrg(ctx workflow.Context, canaryReq *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	l := workflow.GetLogger(ctx)
	wkflowReq := &orgsv1.DeprovisionRequest{
		OrgId:  canaryReq.CanaryId,
		Region: defaultRegion,
	}

	l.Info("deprovisioning org", "request", wkflowReq)
	workflowID, err := w.startWorkflow(ctx, "orgs", "Teardown", wkflowReq)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}

	pollResp, err := w.pollWorkflow(ctx, "orgs", "Teardown", workflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	}
	l.Info("successfully got org response", "response", pollResp)

	return pollResp.Step, nil
}

func (w *wkflow) deprovisionInstall(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	l := workflow.GetLogger(ctx)
	wkflowReq := &installsv1.DeprovisionRequest{
		OrgId:     req.CanaryId,
		AppId:     req.CanaryId,
		InstallId: req.CanaryId,
		AccountSettings: &installsv1.AccountSettings{
			Region:       "us-west-2",
			AwsAccountId: "548377525120",
			AwsRoleArn:   w.cfg.InstallIamRoleArn,
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:    "aws-eks",
			Version: "0.11.1",
		},
	}

	l.Info("deprovisioning install", "request", wkflowReq)
	workflowID, err := w.startWorkflow(ctx, "installs", "Deprovision", wkflowReq)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}

	pollResp, err := w.pollWorkflow(ctx, "installs", "Deprovision", workflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	}
	l.Info("successfully got install response", "response", pollResp)
	return pollResp.Step, nil
}
