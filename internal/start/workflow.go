package start

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-common/shortid"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
	instancesv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/instances/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/powertoolsdev/workers-deployments/internal/start/build"
	"github.com/powertoolsdev/workers-deployments/internal/start/instances"
	"github.com/powertoolsdev/workers-deployments/internal/start/plan"
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

func (w *wkflow) Start(ctx workflow.Context, req *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error) {
	resp := &deploymentsv1.StartResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)

	if err := req.Validate(); err != nil {
		return resp, err
	}

	shortIDs, err := shortid.ParseStrings(req.OrgId, req.AppId, req.DeploymentId)
	if err != nil {
		return resp, fmt.Errorf("unable to parse short IDs: %w", err)
	}
	orgID, appID, deploymentID := shortIDs[0], shortIDs[1], shortIDs[2]

	if err = w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return resp, err
	}

	// run the plan workflow
	planReq := &planv1.PlanRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
		Component:    req.Component,
	}
	planResp, err := execPlan(ctx, w.cfg, planReq)
	if err != nil {
		err = fmt.Errorf("unable to perform build: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	l.Debug(fmt.Sprintf("finished planning %v", planResp))

	// run the build workflow
	bReq := &buildv1.BuildRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
	}
	if !req.PlanOnly {
		var bResp *buildv1.BuildResponse
		bResp, err = execBuild(ctx, w.cfg, bReq)
		if err != nil {
			err = fmt.Errorf("unable to perform build: %w", err)
			w.finishWorkflow(ctx, req, resp, err)
			return resp, err
		}
		l.Debug(fmt.Sprintf("finished build %v", bResp))
	}

	ipReq := &instancesv1.ProvisionRequest{
		OrgId:            orgID,
		AppId:            appID,
		DeploymentId:     deploymentID,
		Plan:             planResp.Plan,
		InstallIds:       req.InstallIds,
		DeploymentPrefix: getS3Prefix(orgID, appID, req.Component.Name, deploymentID),
		PlanOnly:         req.PlanOnly,
	}
	ipResp, err := execProvisionInstances(ctx, w.cfg, ipReq)
	if err != nil {
		err = fmt.Errorf("unable to provision instances: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	resp.WorkflowIds = ipResp.WorkflowIds

	w.finishWorkflow(ctx, req, resp, nil)
	return resp, nil
}

func execPlan(
	ctx workflow.Context,
	cfg workers.Config,
	req *planv1.PlanRequest,
) (*planv1.PlanResponse, error) {
	resp := &planv1.PlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing build workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := plan.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.Plan, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execBuild(
	ctx workflow.Context,
	cfg workers.Config,
	req *buildv1.BuildRequest,
) (*buildv1.BuildResponse, error) {
	resp := &buildv1.BuildResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing build workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := build.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.Build, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execProvisionInstances(
	ctx workflow.Context,
	cfg workers.Config,
	req *instancesv1.ProvisionRequest,
) (*instancesv1.ProvisionResponse, error) {
	resp := &instancesv1.ProvisionResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing build workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := instances.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionInstances, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
