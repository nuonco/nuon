package start

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	deploymentsv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/deployments/v1"
	buildv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/deployments/v1/build/v1"
	instancesv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/deployments/v1/instances/v1"
	workers "github.com/powertoolsdev/mono/services/workers-deployments/internal"
	"github.com/powertoolsdev/mono/services/workers-deployments/internal/start/build"
	"github.com/powertoolsdev/mono/services/workers-deployments/internal/start/instances"
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

	if err := w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return resp, err
	}

	// run the build workflow
	bReq := &buildv1.BuildRequest{
		OrgId:        req.OrgId,
		AppId:        req.AppId,
		DeploymentId: req.DeploymentId,
		Component:    req.Component,
		PlanOnly:     req.PlanOnly,
	}
	bResp, err := execBuild(ctx, w.cfg, bReq)
	if err != nil {
		err = fmt.Errorf("unable to build: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	l.Debug(fmt.Sprintf("finished build %v", bResp))
	resp.PlanRef = bResp.PlanRef

	if req.BuildOnly {
		w.finishWorkflow(ctx, req, resp, nil)
		return resp, nil
	}

	ipReq := &instancesv1.ProvisionRequest{
		OrgId:        req.OrgId,
		AppId:        req.AppId,
		DeploymentId: req.DeploymentId,
		InstallIds:   req.InstallIds,
		Component:    req.Component,
		PlanOnly:     req.PlanOnly,
		BuildPlan:    bResp.PlanRef,
	}
	_, err = execProvisionInstances(ctx, w.cfg, ipReq)
	if err != nil {
		err = fmt.Errorf("unable to provision instances: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	w.finishWorkflow(ctx, req, resp, nil)
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
