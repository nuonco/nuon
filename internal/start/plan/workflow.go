package plan

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
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

func (w *wkflow) Plan(ctx workflow.Context, req *planv1.PlanRequest) (*planv1.PlanResponse, error) {
	resp := &planv1.PlanResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities()

	if err := req.Validate(); err != nil {
		return resp, err
	}

	cpReq := CreatePlanRequest{
		OrgID:                          req.OrgId,
		AppID:                          req.AppId,
		DeploymentID:                   req.DeploymentId,
		DeploymentsBucketPrefix:        getS3Prefix(req),
		DeploymentsBucketAssumeRoleARN: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, req.OrgId),
		Component:                      req.Component,
		Config:                         w.cfg,
	}
	cpResp, err := execCreatePlan(ctx, act, cpReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create plan: %w", err)
	}
	resp.Plan = cpResp.Plan

	l.Debug("successfully created plan for build")
	return resp, nil
}

func execCreatePlan(
	ctx workflow.Context,
	act *Activities,
	req CreatePlanRequest,
) (CreatePlanResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp CreatePlanResponse

	l.Debug("executing create plan activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.CreatePlan, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
