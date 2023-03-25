package workflows

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_orgs.go -source=orgs.go -package=workflows
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
	orgID, err := shortid.ParseString(orgID)
	if err != nil {
		return fmt.Errorf("unable to parse shortid from orgID: %w", err)
	}

	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"started-by": "api",
		},
	}
	args := &orgsv1.SignupRequest{OrgId: orgID, Region: "us-west-2"}

	_, err = o.tc.ExecuteWorkflow(ctx, opts, "Signup", args)
	return err
}

type orgDeprovisionArgs struct {
	OrgID  string `validate:"required" json:"org_id"`
	Region string `validate:"required" json:"region"`
}

func (o *orgWorkflowManager) Deprovision(ctx context.Context, orgID string) error {
	orgID, err := shortid.ParseString(orgID)
	if err != nil {
		return fmt.Errorf("unable to parse shortid from orgID: %w", err)
	}

	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"started-by": "api",
		},
	}
	args := orgDeprovisionArgs{OrgID: orgID, Region: "us-west-2"}

	_, err = o.tc.ExecuteWorkflow(ctx, opts, "Teardown", args)
	return err
}
