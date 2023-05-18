package workflows

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_orgs.go -source=orgs.go -package=workflows
func NewOrgWorkflowManager(tc temporal.Client) *orgWorkflowManager {
	return &orgWorkflowManager{
		tc: tc,
	}
}

type orgWorkflowManager struct {
	tc temporal.Client
}

type OrgWorkflowManager interface {
	Provision(context.Context, string) (string, error)
	Deprovision(context.Context, string) (string, error)
}

var _ OrgWorkflowManager = (*orgWorkflowManager)(nil)

func (o *orgWorkflowManager) Provision(ctx context.Context, orgID string) (string, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"started-by": "api",
		},
	}
	args := &orgsv1.SignupRequest{OrgId: orgID, Region: "us-west-2"}

	workflow, err := o.tc.ExecuteWorkflowInNamespace(ctx, "orgs", opts, "Signup", args)
	if err != nil {
		return "", err
	}

	return workflow.GetID(), err
}

type orgDeprovisionArgs struct {
	OrgID  string `validate:"required" json:"org_id"`
	Region string `validate:"required" json:"region"`
}

func (o *orgWorkflowManager) Deprovision(ctx context.Context, orgID string) (string, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":     orgID,
			"started-by": "api",
		},
	}
	args := orgDeprovisionArgs{OrgID: orgID, Region: "us-west-2"}

	workflow, err := o.tc.ExecuteWorkflowInNamespace(ctx, "orgs", opts, "Teardown", args)
	if err != nil {
		return "", err
	}
	return workflow.GetID(), err
}
