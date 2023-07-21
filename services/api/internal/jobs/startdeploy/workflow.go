package startdeploy

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	connectionsv1 "github.com/powertoolsdev/mono/pkg/types/components/connections/v1"
	buildsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	deploysv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deploys/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	meta "github.com/powertoolsdev/mono/pkg/workflows/meta"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/startdeploy/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func New(v *validator.Validate, cfg Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg Config
}

func (w *wkflow) StartDeploy(ctx workflow.Context, req *jobsv1.StartDeployRequest) (*jobsv1.StartDeployResponse, error) {
	act := activities.NewActivities(nil, nil)
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var idResp activities.GetIDsResponse
	fut := workflow.ExecuteActivity(ctx, act.GetIDs, req.DeployId)
	err := fut.Get(ctx, &idResp)
	if err != nil {
		return nil, fmt.Errorf("unable to get deploy ids: %w", err)
	}

	shrdAct := &sharedactivities.Activities{}

	// poll build workflow future
	pollBuildRequest := &activitiesv1.PollWorkflowRequest{
		Namespace:    "builds",
		WorkflowName: "Build",
		WorkflowId:   idResp.BuildID,
	}

	var pollResp activitiesv1.PollWorkflowResponse
	fut = workflow.ExecuteActivity(ctx, shrdAct.PollWorkflow, pollBuildRequest)
	err = fut.Get(ctx, &pollResp)
	if err != nil {
		return nil, fmt.Errorf("unable to poll for workflow response: %w", err)
	}

	buildResp, err := pollResp.Response.UnmarshalNew()
	if err != nil {
		return nil, fmt.Errorf("error formatting build response from poll: %w", err)
	}

	buildMsg, ok := buildResp.(*buildsv1.BuildResponse)
	if !ok {
		return nil, fmt.Errorf("error creating build response: %t", ok)
	}

	// we have some proto duplications that will be cleaned up in the workflow
	// pkg refactor
	resp := &jobsv1.StartDeployResponse{
		BuildPlan: buildMsg.BuildPlan,
	}

	finishResp := &deploysv1.DeployResponse{
		BuildPlan: buildMsg.BuildPlan,
	}

	// fetch plans from S3 (build)
	// planRef should have S3 path to fetch from
	plan := &planv1.Plan{}
	fut = workflow.ExecuteActivity(ctx, act.FetchBuildPlanJob, resp.BuildPlan)
	if err = fut.Get(ctx, &plan); err != nil {
		return nil, fmt.Errorf("unable to trigger workflow response: %w", err)
	}

	// we need to add the installId and connections to the plan
	switch plan := plan.Actual.(type) {
	case *planv1.Plan_WaypointPlan:
		connections := &connectionsv1.Connections{}
		fut = workflow.ExecuteActivity(ctx, act.AddConnectionsToPlan, plan.WaypointPlan.Component.Id, idResp.InstallID)
		if err = fut.Get(ctx, &connections); err != nil {
			return nil, fmt.Errorf("unable to add connections to plan: %w", err)
		}

		plan.WaypointPlan.Metadata.InstallId = idResp.InstallID
		plan.WaypointPlan.Component.Connections = connections
	}

	// start the executors part of the workflows for syncing and deploying
	if err = w.startWorkflow(ctx, plan, idResp); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		w.finishWorkflow(ctx, plan, finishResp, err)
		return resp, err
	}
	// call child workflow workers-instances
	imageSyncPlanRef, err := w.planAndExec(ctx, plan, planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_SYNC_IMAGE)
	if err != nil {
		err = fmt.Errorf("unable to sync image: %w", err)
		w.finishWorkflow(ctx, plan, finishResp, err)
		return resp, nil
	}

	resp.ImageSyncPlan = imageSyncPlanRef
	finishResp.ImageSyncPlan = imageSyncPlanRef

	// deploy instance
	deployPlanRef, err := w.planAndExec(ctx, plan, planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_DEPLOY)
	if err != nil {
		w.finishWorkflow(ctx, plan, finishResp, err)
		return resp, nil
	}

	resp.DeployPlan = deployPlanRef
	finishResp.DeployPlan = deployPlanRef
	w.finishWorkflow(ctx, plan, finishResp, err)

	//for now we don't update deploy, but if we need to update the Deploy with
	//a status, we can do so here (or in the finishWorkflow function)
	return resp, nil
}

func (w *wkflow) startWorkflow(ctx workflow.Context, plan *planv1.Plan, ids activities.GetIDsResponse) error {
	wpPlan := plan.GetWaypointPlan()
	info := workflow.GetInfo(ctx)

	req := &deploysv1.DeployRequest{
		DeployId: ids.DeployID,
		BuildId:  ids.BuildID,
		OrgId:    wpPlan.Metadata.OrgId,
		AppId:    wpPlan.Metadata.AppId,
	}

	startReq := &sharedv1.StartActivityRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, wpPlan.Metadata.OrgId),
		MetadataBucketPrefix:        prefix.InstancePath(wpPlan.Metadata.OrgId, wpPlan.Metadata.AppId, wpPlan.Component.Id, wpPlan.Metadata.DeploymentId, wpPlan.Metadata.InstallId),
		RequestRef: &sharedv1.RequestRef{
			Request: &sharedv1.RequestRef_DeployRequest{
				DeployRequest: req,
			},
		},
		WorkflowInfo: &sharedv1.WorkflowInfo{
			Id: info.WorkflowExecution.ID,
		},
	}

	if _, err := execStart(ctx, startReq); err != nil {
		return fmt.Errorf("unable to start workflow: %w", err)
	}

	return nil
}

func (w *wkflow) planAndExec(
	ctx workflow.Context,
	buildPlan *planv1.Plan, // this will take BuildPlan from S3
	typ planv1.ComponentInputType) (*planv1.PlanRef, error) {
	waypointPlan := buildPlan.GetWaypointPlan()
	l := workflow.GetLogger(ctx)

	planReq := &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:        waypointPlan.Metadata.OrgId,
				AppId:        waypointPlan.Metadata.AppId,
				InstallId:    waypointPlan.Metadata.InstallId,
				Component:    waypointPlan.Component,
				DeploymentId: waypointPlan.Metadata.DeploymentId,
				Type:         typ,
			},
		},
	}
	planResp, err := execCreatePlan(ctx, planReq)
	if err != nil {
		err = fmt.Errorf("unable to create %s plan: %w", typ, err)
		w.finishWorkflow(ctx, buildPlan, nil, err)
		return nil, err
	}
	l.Debug("finished creating", zap.Any("plan_type", typ))

	execReq := &executev1.ExecutePlanRequest{
		Plan: planResp.Plan,
	}
	_, err = execExecutePlan(ctx, execReq)
	if err != nil {
		err = fmt.Errorf("unable to execute %s plan: %w", typ, err)
		w.finishWorkflow(ctx, buildPlan, nil, err)
		return nil, err
	}
	l.Debug("finished executing %s plan", typ)
	return planResp.Plan, nil
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
) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing execute plan workflow")
	cwo := workflow.ChildWorkflowOptions{
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

// finishWorkflow calls the finish step
//
//nolint:all
func (w *wkflow) finishWorkflow(ctx workflow.Context, req *planv1.Plan, resp *deploysv1.DeployResponse, workflowErr error) {
	plan := req.GetWaypointPlan()
	var err error
	defer func() {
		if err == nil {
			return
		}

		l := workflow.GetLogger(ctx)
		l.Debug("unable to finish workflow: %w", err)
	}()

	status := sharedv1.ResponseStatus_RESPONSE_STATUS_OK
	errMessage := ""
	if workflowErr != nil {
		status = sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR
		errMessage = workflowErr.Error()
	}

	finishReq := &sharedv1.FinishActivityRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, plan.Metadata.OrgId),
		MetadataBucketPrefix:        prefix.InstancePath(plan.Metadata.OrgId, plan.Metadata.AppId, plan.Component.Id, plan.Metadata.DeploymentId, plan.Metadata.InstallId),
		ResponseRef: &sharedv1.ResponseRef{
			Response: &sharedv1.ResponseRef_DeployResponse{
				DeployResponse: resp,
			},
		},
		Status:       status,
		ErrorMessage: errMessage,
	}

	// exec activity
	_, err = execFinish(ctx, finishReq)
	if err != nil {
		err = fmt.Errorf("unable to execute finish activity: %w", err)
	}
}

func execStart(
	ctx workflow.Context,
	req *sharedv1.StartActivityRequest,
) (*sharedv1.StartActivityResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &sharedv1.StartActivityResponse{}

	act := meta.NewStartActivity()
	l.Debug("executing start activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.StartRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func execFinish(
	ctx workflow.Context,
	req *sharedv1.FinishActivityRequest,
) (*sharedv1.FinishActivityResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &sharedv1.FinishActivityResponse{}

	act := meta.NewFinishActivity()
	l.Debug("executing finish activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.FinishRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
