package createorg

import (
	"context"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
)

type activities struct {
	wfc wfc.Client
}

func NewActivities(wfc wfc.Client) *activities {
	return &activities{wfc}
}

type TriggerJobResponse struct {
	WorkflowID string
}

func (a *activities) TriggerOrgProvision(ctx context.Context, orgID string) (*TriggerJobResponse, error) {
	workflowID, err := a.wfc.TriggerOrgSignup(ctx, &orgsv1.SignupRequest{OrgId: orgID})
	if err != nil {
		return nil, err
	}
	return &TriggerJobResponse{
		WorkflowID: workflowID,
	}, nil
}
