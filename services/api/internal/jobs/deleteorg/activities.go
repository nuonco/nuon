package deleteorg

import (
	"context"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
)

type activities struct {
	wfc workflowsclient.Client
}

func NewActivities(workflowsClient workflowsclient.Client) *activities {
	return &activities{
		wfc: workflowsClient,
	}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerOrgDeprovision(ctx context.Context, orgID string) (*TriggerJobResponse, error) {
	req := &orgsv1.TeardownRequest{OrgId: orgID, Region: "us-west-2"}
	workflowID, err := a.wfc.TriggerOrgTeardown(ctx, req)
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
