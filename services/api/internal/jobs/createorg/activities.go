package createorg

import (
	"context"

	pkgWorkflows "github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
)

type activities struct {
	mgr workflows.OrgWorkflowManager
}

func NewActivities(workflowsClient pkgWorkflows.Client) *activities {
	return &activities{
		mgr: workflows.NewOrgWorkflowManager(workflowsClient),
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerOrgProvision(ctx context.Context, orgID string) (*TriggerJobResponse, error) {
	workflowID, err := a.mgr.Provision(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
