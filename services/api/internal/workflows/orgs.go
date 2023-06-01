package workflows

import (
	"context"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	pkgWorkflows "github.com/powertoolsdev/mono/pkg/workflows"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_orgs.go -source=orgs.go -package=workflows
func NewOrgWorkflowManager(workflowsClient pkgWorkflows.Client) *orgWorkflowManager {
	return &orgWorkflowManager{workflowsClient}
}

type orgWorkflowManager struct {
	workflowsClient pkgWorkflows.Client
}

type OrgWorkflowManager interface {
	Provision(context.Context, string) (string, error)
	Deprovision(context.Context, string) (string, error)
}

var _ OrgWorkflowManager = (*orgWorkflowManager)(nil)

func (o *orgWorkflowManager) Provision(ctx context.Context, orgID string) (string, error) {
	args := &orgsv1.SignupRequest{OrgId: orgID, Region: "us-west-2"}

	workflowID, err := o.workflowsClient.TriggerOrgSignup(ctx, args)
	if err != nil {
		return "", err
	}

	return workflowID, err
}

func (o *orgWorkflowManager) Deprovision(ctx context.Context, orgID string) (string, error) {
	workflowID, err := o.workflowsClient.TriggerOrgTeardown(ctx, &orgsv1.TeardownRequest{OrgId: orgID, Region: "us-west-2"})
	if err != nil {
		return "", err
	}
	return workflowID, err
}
