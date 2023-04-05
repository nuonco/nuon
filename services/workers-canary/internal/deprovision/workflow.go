package deprovision

import (
	"time"

	"go.temporal.io/sdk/workflow"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	workers "github.com/powertoolsdev/mono/services/workers-canary/internal"
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

func (w *wkflow) Deprovision(ctx workflow.Context, req *canaryv1.DeprovisionRequest) (*canaryv1.DeprovisionResponse, error) {
	resp := &canaryv1.DeprovisionResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)

	if err := req.Validate(); err != nil {
		return resp, err
	}
	l.Info("executing provision canary")
	return resp, nil
}
