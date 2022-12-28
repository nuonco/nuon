package instances

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-common/shortid"
	instancesv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/instances/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
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

func (w *wkflow) ProvisionInstances(ctx workflow.Context, req *instancesv1.ProvisionRequest) (*instancesv1.ProvisionResponse, error) {
	resp := &instancesv1.ProvisionResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities(workers.Config{})

	if err := req.Validate(); err != nil {
		return resp, err
	}

	// start instance workflows
	for _, installID := range req.InstallIds {
		var installShortID string
		installShortID, err := shortid.ParseString(installID)
		if err != nil {
			return resp, fmt.Errorf("unable to parse short ID for install: %w", err)
		}

		actReq := ProvisionInstanceRequest{
			OrgID:        req.OrgId,
			AppID:        req.AppId,
			DeploymentID: req.DeploymentId,
			InstallID:    installShortID,
			Plan:         req.Plan,
		}

		actResp, err := execProvisionInstanceActivity(ctx, act, actReq)
		if err != nil {
			return resp, fmt.Errorf("unable to execute provision instance activity: %w", err)
		}
		resp.WorkflowIds = append(resp.WorkflowIds, actResp.WorkflowID)
	}

	l.Debug(fmt.Sprintf("starting %d child workflows", len(req.InstallIds)))

	l.Debug("successfully provisioned instances")
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
