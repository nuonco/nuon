package build

import (
	"fmt"
	"time"

	buildsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"github.com/powertoolsdev/mono/services/api/internal"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultActivityTimeout = time.Minute * 1
)

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	return workflow.WithActivityOptions(ctx, activityOpts)
}

type wkflow struct {
	cfg *internal.Config
}

func New(cfg *internal.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) Build(ctx workflow.Context, req *buildsv1.BuildRequest) (*buildsv1.BuildResponse, error) {
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)

	if err := w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return nil, err
	}

	l.Info("creating plan create request")
	createPlanReq, err := execCreatePlanRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan request: %w", err)
	}
	if err = createPlanReq.Validate(); err != nil {
		return nil, fmt.Errorf("unable to create a plan request: %w", err)
	}

	l.Info("creating plan")
	createPlanResp, err := execCreatePlan(ctx, createPlanReq, req.BuildId)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	res := &buildsv1.BuildResponse{
		BuildPlan: createPlanResp.Plan,
	}

	l.Info("executing plan")
	_, err = execExecutePlan(ctx, &executev1.ExecutePlanRequest{
		Plan: createPlanResp.Plan,
	}, req.BuildId)
	if err != nil {
		w.finishWorkflow(ctx, req, res, err)
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	w.finishWorkflow(ctx, req, res, nil)
	return res, nil
}

// execCreatePlanRequest returns a plan request that can be passed to a workflow
func execCreatePlanRequest(
	ctx workflow.Context,
	req *buildsv1.BuildRequest,
) (*planv1.CreatePlanRequest, error) {
	resp := &planv1.CreatePlanRequest{}
	l := workflow.GetLogger(ctx)

	// This call with nil is kind of a hacky way to get references to the activity methods,
	// but is not really for code execution since the activity invocations happens
	// over the wire and we can't serialize anything other than pure data arguments
	a := activities.New(nil, "", "")
	opts := workflow.ActivityOptions{ScheduleToCloseTimeout: time.Second * 5}
	ctx = workflow.WithActivityOptions(ctx, opts)

	l.Info("executing create plan request activity")
	fut := workflow.ExecuteActivity(ctx, a.CreatePlanRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execCreatePlan(
	ctx workflow.Context,
	req *planv1.CreatePlanRequest,
	buildID string,
) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing create plan workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-create-plan", buildID),
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		TaskQueue:                wfc.ExecutorsTaskQueue,
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
	buildID string,
) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing execute plan workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-execute-plan", buildID),
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		TaskQueue:                wfc.ExecutorsTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// startWorkflow is a utility method for executing the StartStartRequest activity
func (w *wkflow) startWorkflow(ctx workflow.Context, req *buildsv1.BuildRequest) error {
	info := workflow.GetInfo(ctx)
	prefix := getS3PrefixFromRequest(req)

	startReq := &sharedv1.StartActivityRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, req.OrgId),
		MetadataBucketPrefix:        prefix,
		RequestRef:                  metaRequestFromReq(req),
		WorkflowInfo: &sharedv1.WorkflowInfo{
			Id: info.WorkflowExecution.ID,
		},
	}

	act := activities.New(nil, "", "")
	if _, err := execStart(ctx, act, startReq); err != nil {
		return fmt.Errorf("unable to start workflow: %w", err)
	}

	return nil
}

func execStart(
	ctx workflow.Context,
	act *activities.Activities,
	req *sharedv1.StartActivityRequest,
) (*sharedv1.StartActivityResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &sharedv1.StartActivityResponse{}

	l.Debug("executing start activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.StartStartRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func metaRequestFromReq(req *buildsv1.BuildRequest) *sharedv1.RequestRef {
	return &sharedv1.RequestRef{
		Request: &sharedv1.RequestRef_BuildRequest{
			BuildRequest: req,
		},
	}
}

// finishWorkflow is a utility method for executing the FinishStartRequest activity
func (w *wkflow) finishWorkflow(ctx workflow.Context, req *buildsv1.BuildRequest, resp *buildsv1.BuildResponse, workflowErr error) {
	var err error
	defer func() {
		if err == nil {
			return
		}

		l := workflow.GetLogger(ctx)
		l.Debug("unable to finish workflow: %w", err)
	}()

	prefix := getS3PrefixFromRequest(req)

	status := sharedv1.ResponseStatus_RESPONSE_STATUS_OK
	errMessage := ""
	if workflowErr != nil {
		status = sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR
		errMessage = workflowErr.Error()
	}

	finishReq := &sharedv1.FinishActivityRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, req.OrgId),
		MetadataBucketPrefix:        prefix,
		ResponseRef:                 metaResponseFromResponse(resp),
		Status:                      status,
		ErrorMessage:                errMessage,
	}

	// exec activity
	act := activities.New(nil, "", "")
	_, err = execFinish(ctx, act, finishReq)
	if err != nil {
		err = fmt.Errorf("unable to execute finish activity: %w", err)
	}
}

func execFinish(
	ctx workflow.Context,
	act *activities.Activities,
	req *sharedv1.FinishActivityRequest,
) (*sharedv1.FinishActivityResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &sharedv1.FinishActivityResponse{}

	l.Debug("executing finish activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.FinishStartRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func metaResponseFromResponse(resp *buildsv1.BuildResponse) *sharedv1.ResponseRef {
	return &sharedv1.ResponseRef{
		Response: &sharedv1.ResponseRef_BuildResponse{
			BuildResponse: resp,
		},
	}
}

func getS3PrefixFromRequest(req *buildsv1.BuildRequest) string {
	// Providing BuildId here instead of DeploymentId.
	// Code and services downstream of this still require a DeploymentId,
	// so we need to continue setting that for now.
	return getS3Prefix(req.OrgId, req.AppId, req.ComponentId, req.BuildId)
}

// getS3Prefix returns the prefix to be used for the plan and it's encompassed files
func getS3Prefix(orgID, appID, componentID, deploymentID string) string {
	return prefix.BuildPath(orgID, appID, componentID, deploymentID)
}
