package workflows

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type provisionStep struct {
	name string
	fn   func(workflow.Context, *canaryv1.ProvisionRequest) (*canaryv1.Step, error)
}

func (w *wkflow) Provision(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.ProvisionResponse, error) {
	resp := &canaryv1.ProvisionResponse{
		Steps:    make([]*canaryv1.Step, 0),
		CanaryId: req.CanaryId,
	}

	l := workflow.GetLogger(ctx)
	l.Info("provisioning canary", "id", req.CanaryId)

	ensureCanaryID := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		if req.CanaryId != "" {
			return req.CanaryId
		}

		newCanaryID := domains.NewCanaryID() //prefix=def
		return newCanaryID
	})

	var canaryID string
	if err := ensureCanaryID.Get(&canaryID); err != nil {
		return resp, fmt.Errorf("unable to get canary ID: %w", err)
	}
	req.CanaryId = canaryID

	if err := req.Validate(); err != nil {
		return resp, err
	}

	w.sendNotification(ctx, notificationTypeProvisionStart, req.CanaryId, nil)
	steps := []provisionStep{
		{
			"org",
			w.provisionOrg,
		},
		{
			"app",
			w.provisionApp,
		},
		{
			"install",
			w.provisionInstall,
		},
		//{
		//"docker-pull-deployment",
		//w.provisionDeployment,
		//},
	}

	for _, step := range steps {
		stepResp, err := step.fn(ctx, req)
		if err != nil {
			err = fmt.Errorf("unable to provision %s: %w", step.name, err)
			w.sendNotification(ctx, notificationTypeProvisionError, req.CanaryId, err)

			if depErr := w.execProvisionDeprovision(ctx, req); depErr != nil {
				l.Info("unable to start deprovision after error", zap.Error(depErr))
			}

			return resp, err
		}

		resp.Steps = append(resp.Steps, stepResp)
		l.Info("successfully executed %s step", step.name)
	}

	w.sendNotification(ctx, notificationTypeProvisionSuccess, req.CanaryId, nil)
	if err := w.execProvisionDeprovision(ctx, req); err != nil {
		err = fmt.Errorf("unable to start deprovision workflow")
		w.sendNotification(ctx, notificationTypeProvisionError, req.CanaryId, err)
		return resp, err
	}

	return resp, nil
}

func (w *wkflow) provisionOrg(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	l := workflow.GetLogger(ctx)
	wkflowReq := &orgsv1.ProvisionRequest{
		OrgId:  req.CanaryId,
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

func (w *wkflow) provisionApp(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	l := workflow.GetLogger(ctx)
	wkflowReq := &appsv1.ProvisionRequest{
		OrgId: req.CanaryId,
		AppId: req.CanaryId,
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

func (w *wkflow) provisionInstall(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	l := workflow.GetLogger(ctx)
	wkflowReq := &installsv1.ProvisionRequest{
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

	l.Info("provisioning install", "request", wkflowReq)
	workflowID, err := w.startWorkflow(ctx, "installs", "Provision", wkflowReq)
	if err != nil {
		return nil, fmt.Errorf("unable to start workflow: %w", err)
	}

	pollResp, err := w.pollWorkflow(ctx, "installs", "Provision", workflowID)
	if err != nil {
		return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	}
	l.Info("successfully got install response", "response", pollResp)
	return pollResp.Step, nil
}

//nolint:all
func (w *wkflow) provisionDeployment(ctx workflow.Context, req *canaryv1.ProvisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): build out actual deployment request
	return nil, fmt.Errorf("not implemented")
}

func (w *wkflow) execProvisionDeprovision(ctx workflow.Context, req *canaryv1.ProvisionRequest) error {
	if !req.Deprovision {
		return nil
	}
	l := workflow.GetLogger(ctx)

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Hour * 24,
		WorkflowTaskTimeout:      time.Hour,
		TaskQueue:                workflowsclient.DefaultTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "Deprovision", &canaryv1.DeprovisionRequest{
		CanaryId: req.CanaryId,
	})

	var resp canaryv1.DeprovisionResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}
	l.Debug("deprovision response", "response", &resp)
	return nil
}
