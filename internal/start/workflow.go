package start

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-waypoint"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/powertoolsdev/workers-deployments/internal/start/build"
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
	act := NewActivities(workers.Config{})

	if err := req.Validate(); err != nil {
		return resp, err
	}

	orgID, err := shortid.ParseString(req.OrgId)
	if err != nil {
		return resp, fmt.Errorf("unable to get short org ID: %w", err)
	}
	appID, err := shortid.ParseString(req.AppId)
	if err != nil {
		return resp, fmt.Errorf("unable to get short app ID: %w", err)
	}

	deploymentID, err := shortid.ParseString(req.DeploymentId)
	if err != nil {
		return resp, fmt.Errorf("unable to get short deployment ID: %w", err)
	}

	// run the build workflow
	bReq := &buildv1.BuildRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
	}
	bResp, err := execBuild(ctx, w.cfg, bReq)
	if err != nil {
		return resp, fmt.Errorf("unable to perform build: %w", err)
	}
	l.Debug(fmt.Sprintf("finished build %v", bResp))

	// start instance workflows
	for _, installID := range req.InstallIds {
		installShortID, err := shortid.ParseString(installID)
		if err != nil {
			return resp, err
		}

		actReq := ProvisionInstanceRequest{
			OrgID:        orgID,
			AppID:        appID,
			DeploymentID: deploymentID,
			InstallID:    installShortID,
			Component: waypoint.Component{
				Name:              "mario",
				ID:                "mario",
				ContainerImageURL: "kennethreitz/httpbin",
				Type:              "public",
			},
		}

		actResp, err := execProvisionInstanceActivity(ctx, act, actReq)
		if err != nil {
			return resp, err
		}
		resp.WorkflowIds = append(resp.WorkflowIds, actResp.WorkflowID)
	}

	l.Debug(fmt.Sprintf("starting %d child workflows", len(req.InstallIds)))
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

func execProvisionInstanceActivity(
	ctx workflow.Context,
	act *Activities,
	req ProvisionInstanceRequest,
) (ProvisionInstanceResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := ProvisionInstanceResponse{}

	l.Debug("executing provision instance activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.ProvisionInstance, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
