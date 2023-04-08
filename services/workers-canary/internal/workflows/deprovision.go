package workflows

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"go.temporal.io/sdk/workflow"
)

type deprovisionStep struct {
	name string
	fn   func(workflow.Context, *canaryv1.DeprovisionRequest) (*canaryv1.Step, error)
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

	w.sendNotification(ctx, notificationTypeDeprovisionStart, req.CanaryId, nil)
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
		stepResp, err := step.fn(ctx, req)
		if err != nil {
			err = fmt.Errorf("unable to provision %s: %w", step.name, err)
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
	wkflowReq := &orgsv1.SignupRequest{
		OrgId:  canaryReq.CanaryId,
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

func (w *wkflow) deprovisionApp(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): implement deprovision app workflow
	return nil, fmt.Errorf("not implemented")
}

//nolint:all
func (w *wkflow) deprovisionInstall(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): build out actual install request
	return nil, fmt.Errorf("not implemented")
	//l := workflow.GetLogger(ctx)
	//wkflowReq := &installsv1.DeprovisionRequest{
	//OrgId:	   canaryID,
	//AppId:	   canaryID,
	//InstallId: canaryID,
	//}

	//l.Info("provisioning app", "request", wkflowReq)
	//workflowID, err := w.startWorkflow(ctx, "apps", "Deprovision", wkflowReq)
	//if err != nil {
	//return nil, fmt.Errorf("unable to start workflow: %w", err)
	//}

	//pollResp, err := w.pollWorkflow(ctx, "apps", "Deprovision", workflowID)
	//if err != nil {
	//return nil, fmt.Errorf("unable to get finished workflow: %w", err)
	//}
	//l.Info("successfully got app response", "response", pollResp)
	//return nil, nil
}

//nolint:all
func (w *wkflow) deprovisionDeployment(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.Step, error) {
	// TODO(jm): build out actual deployment request
	return nil, fmt.Errorf("not implemented")
}
