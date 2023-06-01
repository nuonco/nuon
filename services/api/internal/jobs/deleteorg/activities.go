package deleteorg

import (
	"context"

	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
)

type activities struct {
	mgr workflows.OrgWorkflowManager
}

func NewActivities(workflowsClient workflowsclient.Client) *activities {
	return &activities{
		mgr: workflows.NewOrgWorkflowManager(workflowsClient),
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
