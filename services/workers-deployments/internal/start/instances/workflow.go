package instances

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	instancesv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/deployments/v1/instances/v1"
	provisionv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/instances/v1"
	workers "github.com/powertoolsdev/mono/services/workers-deployments/internal"
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
		actReq := &provisionv1.ProvisionRequest{
			OrgId:        req.OrgId,
			AppId:        req.AppId,
			DeploymentId: req.DeploymentId,
			InstallId:    installID,
			Component:    req.Component,
			PlanOnly:     req.PlanOnly,
			BuildPlan:    req.BuildPlan,
		}

		_, err := execProvisionInstanceActivity(ctx, act, actReq)
		if err != nil {
			l.Error("failed to provision instance", zap.Error(err))
		}
	}

	l.Debug(fmt.Sprintf("starting %d child workflows", len(req.InstallIds)))

	l.Debug("successfully provisioned instances")
	return resp, nil
}

func execProvisionInstanceActivity(
	ctx workflow.Context,
	act *Activities,
	req *provisionv1.ProvisionRequest,
) (*provisionv1.ProvisionResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &provisionv1.ProvisionResponse{}

	l.Debug("executing provision instance activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.ProvisionInstance, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
