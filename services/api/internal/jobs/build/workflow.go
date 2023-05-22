package build

import (
	"time"

	apibuildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type wkflow struct {
	cfg Config
}

func New(cfg Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) Build(wfctx workflow.Context, req *apibuildv1.StartBuildRequest, db *gorm.DB) (*planv1.PlanRef, error) {
	planRef := &planv1.PlanRef{}
	l := workflow.GetLogger(wfctx)

	createPlanReq := &planv1.CreatePlanRequest{}
	// This call with nil is kind of a hacky way to get references to the activity methods,
	// but is not really for code execution since the activity invocations happens
	// over the wire and we can't serialize anything other than pure data arguments
	a := activities.New(nil, "", "")
	activityOpts := workflow.ActivityOptions{ScheduleToCloseTimeout: time.Second * 5}

	wfctx = workflow.WithActivityOptions(wfctx, activityOpts)
	fut := workflow.ExecuteActivity(wfctx, a.CreatePlanRequest, req)
	if err := fut.Get(wfctx, &createPlanReq); err != nil {
		return nil, err
	}

	// TODO(jm): handle noop builds better
	if createPlanReq.GetComponent().Component.GetBuildCfg().GetNoop() != nil {
		return planRef, nil
	}
	l.Debug("executing create plan workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 1,
		TaskQueue:                workflows.ExecutorsTaskQueue,
	}
	ctx := workflow.WithChildOptions(wfctx, cwo)

	fut = workflow.ExecuteChildWorkflow(ctx, "CreatePlan", createPlanReq)
	if err := fut.Get(ctx, &planRef); err != nil {
		return nil, err
	}
	l.Debug("successfully created plan: %v", planRef)

	l.Debug("executing execute plan workflow")
	cwo = workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		TaskQueue:                workflows.ExecutorsTaskQueue,
	}
	ctx = workflow.WithChildOptions(wfctx, cwo)

	execPlanResp := &executev1.ExecutePlanResponse{}
	fut = workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", executev1.ExecutePlanRequest{Plan: planRef})
	if err := fut.Get(ctx, &execPlanResp); err != nil {
		return nil, err
	}
	l.Debug("successfully executed: %v", execPlanResp, zap.Any("outputs", execPlanResp.Outputs))

	artifact := &models.Artifact{}
	// This call with nil is kind of a hacky way to get references to the activity methods,
	// but is not really for code execution since the activity invocations happens
	// over the wire and we can't serialize anything other than pure data arguments
	fut = workflow.ExecuteActivity(wfctx, a.InsertArtifact, artifact)
	if err := fut.Get(wfctx, &artifact); err != nil {
		return nil, err
	}

	// TODO we probably want to add the new artifactid to the response struct
	return planRef, nil
}
