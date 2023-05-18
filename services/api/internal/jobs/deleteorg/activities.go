package deleteorg

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
)

type activities struct {
	mgr workflows.OrgWorkflowManager
}

func NewActivities(tc temporal.Client) *activities {
	return &activities{
		mgr: workflows.NewOrgWorkflowManager(tc),
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerOrgDeprovision(ctx context.Context, orgID string) (*TriggerJobResponse, error) {
	workflowID, err := a.mgr.Deprovision(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
