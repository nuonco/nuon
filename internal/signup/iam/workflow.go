package iam

import (
	"fmt"
	"time"

	iamv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1/iam/v1"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultActivityTimeout time.Duration = time.Second * 10
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

// ProvisionIAM is a workflow that creates org specific IAM roles in the designated orgs IAM account
func (w wkflow) ProvisionIAM(ctx workflow.Context, req *iamv1.ProvisionIAMRequest) (*iamv1.ProvisionIAMResponse, error) {
	resp := &iamv1.ProvisionIAMResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// get waypoint server cookie
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	l.Debug("creating deployments bucket IAM role")
	dbrRequest := CreateDeploymentsBucketRoleRequest{}
	_, err := execCreateDeploymentsBucketRoleRequest(ctx, act, dbrRequest)
	if err != nil {
		err = fmt.Errorf("failed to create deployments bucket role: %w", err)
		l.Debug(err.Error())
		return resp, err
	}

	return resp, nil
}

func execCreateDeploymentsBucketRoleRequest(
	ctx workflow.Context,
	act *Activities,
	req CreateDeploymentsBucketRoleRequest,
) (CreateDeploymentsBucketRoleResponse, error) {
	var resp CreateDeploymentsBucketRoleResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing ping waypoint server activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateDeploymentsBucketRole, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
