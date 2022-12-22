package plan

import (
	"time"

	"go.temporal.io/sdk/workflow"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
)

//nolint:all
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
	//ctx = configureActivityOptions(ctx)
	//act := NewActivities()

	if err := req.Validate(); err != nil {
		return resp, err
	}

	l.Debug("successfully created plan for build")
	return resp, nil
}
