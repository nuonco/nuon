package start

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-waypoint"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/plan/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/powertoolsdev/workers-deployments/internal/start/build"
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

// parseShortIDs: parse the ids and return any error if found
func parseShortIDs(ids ...string) ([]string, error) {
	shortIDs := make([]string, len(ids))

	for idx, id := range ids {
		parsedID, err := shortid.ParseString(id)
		if err != nil {
			return nil, fmt.Errorf("unable to parse short id: %v: %w", idx, err)
		}
		shortIDs[idx] = parsedID
	}

	return shortIDs, nil
}

// finishWorkflow is a "best effort" finisher for the workflow by executing an activity to emit a response into s3
func (w *wkflow) finishWorkflow(ctx workflow.Context, req *deploymentsv1.StartRequest, resp *deploymentsv1.StartResponse, workflowErr error) {
	l := workflow.GetLogger(ctx)
	status := sharedv1.ResponseStatus_RESPONSE_STATUS_OK
	errMessage := ""
	if workflowErr != nil {
		status = sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR
		errMessage = workflowErr.Error()
	}

	shortIDs, err := parseShortIDs(req.OrgId, req.AppId, req.DeploymentId)
	if err != nil {
		l.Debug("error parsing shortIDs: %w", err)
		return
	}

	prefix := getS3Prefix(shortIDs[0], shortIDs[1], req.Component.Name, shortIDs[2])
	finishReq := FinishRequest{
		DeploymentsBucket:              w.cfg.DeploymentsBucket,
		DeploymentsBucketAssumeRoleARN: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, shortIDs[0]),
		DeploymentsBucketPrefix:        prefix,
		Response:                       resp,
		ResponseStatus:                 status,
		ErrorMessage:                   errMessage,
	}

	// exec activity
	act := NewActivities(workers.Config{})
	_, err = execFinish(ctx, act, finishReq)
	if err != nil {
		l.Debug("unable to execute finish activity: %w", err)
		return
	}
}

//nolint:funlen
func (w *wkflow) Start(ctx workflow.Context, req *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error) {
	resp := &deploymentsv1.StartResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities(workers.Config{})

	if err := req.Validate(); err != nil {
		return resp, err
	}

	shortIDs, err := parseShortIDs(req.OrgId, req.AppId, req.DeploymentId)
	if err != nil {
		return resp, fmt.Errorf("unable to parse short IDs: %w", err)
	}
	orgID, appID, deploymentID := shortIDs[0], shortIDs[1], shortIDs[2]

	prefix := getS3Prefix(orgID, appID, req.Component.Name, deploymentID)
	info := workflow.GetInfo(ctx)
	startReq := StartRequest{
		DeploymentsBucket:              w.cfg.DeploymentsBucket,
		DeploymentsBucketAssumeRoleARN: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, orgID),
		DeploymentsBucketPrefix:        prefix,
		Request:                        req,
		WorkflowInfo: WorkflowInfo{
			ID: info.WorkflowExecution.ID,
		},
	}
	if _, err = execStart(ctx, act, startReq); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, nil
	}

	// run the plan workflow
	planReq := &planv1.PlanRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
	}
	planResp, err := execPlan(ctx, w.cfg, planReq)
	if err != nil {
		err = fmt.Errorf("unable to perform build: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	l.Debug(fmt.Sprintf("finished planning %v", planResp))

	if req.PlanOnly {
		w.finishWorkflow(ctx, req, resp, nil)
		return resp, nil
	}

	// run the build workflow
	bReq := &buildv1.BuildRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
	}
	bResp, err := execBuild(ctx, w.cfg, bReq)
	if err != nil {
		err = fmt.Errorf("unable to perform build: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	l.Debug(fmt.Sprintf("finished build %v", bResp))

	// start instance workflows
	for _, installID := range req.InstallIds {
		var installShortID string
		installShortID, err = shortid.ParseString(installID)
		if err != nil {
			err = fmt.Errorf("unable to parse short ID for install: %w", err)
			w.finishWorkflow(ctx, req, resp, err)
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
			err = fmt.Errorf("unable to execute provision instance activity: %w", err)
			w.finishWorkflow(ctx, req, resp, err)
			return resp, err
		}
		resp.WorkflowIds = append(resp.WorkflowIds, actResp.WorkflowID)
	}

	l.Debug(fmt.Sprintf("starting %d child workflows", len(req.InstallIds)))
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

func execStart(
	ctx workflow.Context,
	act *Activities,
	req StartRequest,
) (StartResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := StartResponse{}

	l.Debug("executing start activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.StartRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func execFinish(
	ctx workflow.Context,
	act *Activities,
	req FinishRequest,
) (FinishResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := FinishResponse{}

	l.Debug("executing finish activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.FinishRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
