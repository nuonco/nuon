package workflows

import (
	"context"

	orgsv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1"
	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_orgs.go -source=orgs.go -package=workflows
const orgTaskQueue string = "org"

func NewOrgWorkflowManager(tc temporalClient) *orgWorkflowManager {
	return &orgWorkflowManager{
		tc: tc,
	}
}

type orgWorkflowManager struct {
	tc temporalClient
}

type OrgWorkflowManager interface {
	Provision(context.Context, string) error
	Deprovision(context.Context, string) error
}

var _ OrgWorkflowManager = (*orgWorkflowManager)(nil)

func (o *orgWorkflowManager) Provision(ctx context.Context, orgID string) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: orgTaskQueue,
	}
	args := &orgsv1.SignupRequest{OrgId: orgID, Region: "us-west-2"}

	_, err := o.tc.ExecuteWorkflow(ctx, opts, "Signup", args)
	return err
}

type orgDeprovisionArgs struct {
	OrgID  string `validate:"required" json:"org_id"`
	Region string `validate:"required" json:"region"`
}

func (o *orgWorkflowManager) Deprovision(ctx context.Context, orgID string) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: orgTaskQueue,
	}
	args := orgDeprovisionArgs{OrgID: orgID, Region: "us-west-2"}

	_, err := o.tc.ExecuteWorkflow(ctx, opts, "Teardown", args)
	return err
}
