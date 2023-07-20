package deleteorg

import (
	"context"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
)

type activities struct {
	wfc wfc.Client
}

func NewActivities(workflowsClient wfc.Client) *activities {
	return &activities{
		wfc: workflowsClient,
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerOrgDeprovision(ctx context.Context, orgID string) (*TriggerJobResponse, error) {
	req := &orgsv1.DeprovisionRequest{OrgId: orgID, Region: "us-west-2"}
	workflowID, err := a.wfc.TriggerOrgTeardown(ctx, req)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
