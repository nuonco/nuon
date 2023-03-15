package kms

import (
	"fmt"
	"time"

	kmsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/kms/v1"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultActivityTimeout time.Duration = time.Second * 10
)

func defaultIAMPath(orgID string) string {
	return fmt.Sprintf("/orgs/%s/", orgID)
}

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
//
//nolint:all
func (w wkflow) ProvisionKMS(ctx workflow.Context, req *kmsv1.ProvisionKMSRequest) (*kmsv1.ProvisionKMSResponse, error) {
	resp := &kmsv1.ProvisionKMSResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)
	// TODO(jm): implement kms functionality
	return resp, nil
}

//nolint:all
func execCreateKMSKey(
	ctx workflow.Context,
	act *Activities,
	req CreateKMSKeyPolicyRequest,
) (CreateKMSKeyPolicyResponse, error) {
	var resp CreateKMSKeyPolicyResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing create iam role activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateKMSKey, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

//nolint:all
func execCreateKMSKeyPolicy(
	ctx workflow.Context,
	act *Activities,
	req CreateKMSKeyPolicyRequest,
) error {
	l := workflow.GetLogger(ctx)

	l.Debug("executing create kms key policy activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateKMSKeyPolicy, req)

	var resp CreateKMSKeyPolicyResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}

	return nil
}
