package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	instancesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/instances/v1"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	"github.com/powertoolsdev/mono/pkg/workflows"
	workers "github.com/powertoolsdev/mono/services/workers-instances/internal"
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

func (w *wkflow) planAndExec(ctx workflow.Context, req *instancesv1.ProvisionRequest, typ planv1.PlanType) (*planv1.PlanRef, error) {
	l := workflow.GetLogger(ctx)

	planReq := &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.Component{
				OrgId:        req.OrgId,
				AppId:        req.AppId,
				DeploymentId: req.DeploymentId,
				InstallId:    req.InstallId,
				Component:    req.Component,
			},
		},
		Type: typ,
	}
	planResp, err := execCreatePlan(ctx, planReq)
	if err != nil {
		err = fmt.Errorf("unable to create %s plan: %w", typ, err)
		w.finishWorkflow(ctx, req, nil, err)
		return nil, err
	}
	l.Debug("finished creating", zap.Any("plan_type", typ))

	if req.PlanOnly {
		l.Debug("plan only enabled, skipping execution of", zap.Any("plan_type", typ))
		return planResp.Plan, nil
	}

	execReq := &executev1.ExecutePlanRequest{
		Plan: planResp.Plan,
	}
	_, err = execExecutePlan(ctx, execReq)
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
	act := NewActivities(nil, nil)

	if err := w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	imageSyncPlanRef, err := w.planAndExec(ctx, req, planv1.PlanType_PLAN_TYPE_WAYPOINT_SYNC_IMAGE)
	if err != nil {
		err = fmt.Errorf("unable to sync image: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, nil
	}
	resp.ImageSyncPlan = imageSyncPlanRef
	l.Debug("successfully deployed", zap.Any("plan_only", req.PlanOnly))

	deployPlanRef, err := w.planAndExec(ctx, req, planv1.PlanType_PLAN_TYPE_WAYPOINT_DEPLOY)
	if err != nil {
		return resp, nil
	}
	resp.DeployPlan = deployPlanRef
	l.Debug("successfully deployed", zap.Any("plan_only", req.PlanOnly))

	if !req.PlanOnly {
		shnReq := SendHostnameNotificationRequest{
			OrgID:                req.OrgId,
			TokenSecretNamespace: w.cfg.WaypointTokenSecretNamespace,
			OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId),
			InstallID:            req.InstallId,
			ComponentID:          req.Component.Id,
		}
		shnResp, err := execSendHostnameNotification(ctx, act, shnReq)
		if err != nil {
			return resp, err
		}
		resp.Hostname = shnResp.Hostname
		l.Debug("successfully sent hostname notification: ", shnResp)
	}

	w.finishWorkflow(ctx, req, resp, nil)
	l.Debug("successfully wrote response")
	return resp, nil
}

func execCreatePlan(
	ctx workflow.Context,
	req *planv1.CreatePlanRequest,
) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing create plan workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		TaskQueue:                workflows.DefaultTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "CreatePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execExecutePlan(
	ctx workflow.Context,
	req *executev1.ExecutePlanRequest,
) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing execute plan workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		TaskQueue:                workflows.DefaultTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", req)
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
