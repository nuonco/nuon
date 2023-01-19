package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-waypoint"
	deploymentplanv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
	instancesv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1/execute/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-instances/internal"
	"github.com/powertoolsdev/workers-instances/internal/provision/execute"
	"github.com/powertoolsdev/workers-instances/internal/provision/plan"
)

const (
	defaultActivityTimeout = time.Second * 5
	defaultDeployTimeout   = time.Minute * 15
)

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	return workflow.WithActivityOptions(ctx, activityOpts)
}

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) planAndExec(ctx workflow.Context, req *instancesv1.ProvisionRequest, typ planv1.PlanType) (*deploymentplanv1.PlanRef, error) {
	l := workflow.GetLogger(ctx)

	planReq := &planv1.CreatePlanRequest{
		BuildPlan: req.BuildPlan,
		InstallId: req.InstallId,
		Type:      typ,
	}
	planResp, err := execCreatePlan(ctx, w.cfg, planReq)
	if err != nil {
		err = fmt.Errorf("unable to create %s plan: %w", typ, err)
		w.finishWorkflow(ctx, req, nil, err)
		return nil, err
	}
	l.Debug("finished creating %s plan", typ)

	execReq := &executev1.ExecuteRequest{
		Plan: planResp.Plan,
	}
	_, err = execExecutePlan(ctx, w.cfg, execReq)
	if err != nil {
		err = fmt.Errorf("unable to execute %s plan: %w", typ, err)
		w.finishWorkflow(ctx, req, nil, err)
		return nil, err
	}
	l.Debug("finished executing %s plan", typ)
	return planResp.Plan, nil
}

func (w *wkflow) Provision(ctx workflow.Context, req *instancesv1.ProvisionRequest) (*instancesv1.ProvisionResponse, error) {
	resp := &instancesv1.ProvisionResponse{
		BuildPlan: req.BuildPlan,
	}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities(nil)

	if err := w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return resp, err
	}

	imageSyncPlanRef, err := w.planAndExec(ctx, req, planv1.PlanType_PLAN_TYPE_SYNC_IMAGE)
	if err != nil {
		return resp, nil
	}
	resp.ImageSyncPlan = imageSyncPlanRef
	l.Debug("successfully synced image")

	deployPlanRef, err := w.planAndExec(ctx, req, planv1.PlanType_PLAN_TYPE_DEPLOY)
	if err != nil {
		return resp, nil
	}
	resp.DeployPlan = deployPlanRef
	l.Debug("successfully deployed")

	// TODO(jm): change this from sending the hostname to slack to just writing it into the response
	shnReq := SendHostnameNotificationRequest{
		OrgID:                "todo-org-id",
		TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, "todo-org-id"),
		InstallID:            req.InstallId,
		AppID:                "todo-app-id",
	}
	shnResp, err := execSendHostnameNotification(ctx, act, shnReq)
	if err != nil {
		return resp, err
	}
	l.Debug("successfully sent hostname notification: ", shnResp)
	resp.Hostname = shnResp.Hostname

	return resp, nil
}

func execCreatePlan(
	ctx workflow.Context,
	cfg workers.Config,
	req *planv1.CreatePlanRequest,
) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing create plan child workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := plan.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.CreatePlan, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execExecutePlan(
	ctx workflow.Context,
	cfg workers.Config,
	req *executev1.ExecuteRequest,
) (*executev1.ExecuteResponse, error) {
	resp := &executev1.ExecuteResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing plan execution workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := execute.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ExecutePlan, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execSendHostnameNotification(
	ctx workflow.Context,
	act *Activities,
	req SendHostnameNotificationRequest,
) (SendHostnameNotificationResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp SendHostnameNotificationResponse

	l.Debug("executing send hostname notification", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.SendHostnameNotification, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
