package deleteorg

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

type TriggerJobResponse struct{}

func (a *activities) TriggerOrgJob(ctx context.Context, orgID string) (*TriggerJobResponse, error) {
	err := a.mgr.Deprovision(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{}, nil
}
