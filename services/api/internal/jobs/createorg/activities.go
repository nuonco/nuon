package createorg

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/workflows"
	tclient "go.temporal.io/sdk/client"
)

type activities struct {
	mgr workflows.OrgWorkflowManager
}

func NewActivities(tc tclient.Client) *activities {
	return &activities{
		mgr: workflows.NewOrgWorkflowManager(tc),
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerJob(ctx context.Context, orgID string) (*TriggerJobResponse, error) {
	workflowID, err := a.mgr.Provision(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
