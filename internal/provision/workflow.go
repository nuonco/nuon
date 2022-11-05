package provision

import (
	"fmt"
	"time"

	workers "github.com/powertoolsdev/template-go-workers/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type ProvisionRequest struct {
	OrgID string `json:"org_id" validate:"required"`
}

type ProvisionResponse struct{}

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w Workflow) Provision(ctx workflow.Context, req ProvisionRequest) (ProvisionResponse, error) {
	resp := ProvisionResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	act := NewActivities()
	dReq := DoRequest{}
	dResp, err := execDo(ctx, act, dReq)
	if err != nil {
		return resp, fmt.Errorf("failed to execute do: %w", err)
	}
	l.Debug("finished do: %s", dResp)

	l.Debug("finished signup", "response", resp)
	return resp, nil
}

func execDo(
	ctx workflow.Context,
	act *Activities,
	req DoRequest,
) (DoResponse, error) {
	var resp DoResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing do request activity")
	fut := workflow.ExecuteActivity(ctx, act.Do, req)

	if err := fut.Get(ctx, &resp); err != nil {
		l.Error("error executing do: %s", err)
	}

	return resp, nil
}
