package services

import (
	"context"
	"fmt"

	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *service) GetInfo(ctx context.Context, orgID string) (*orgsv1.GetInfoResponse, error) {
	workflowResp, err := s.WorkflowsRepo.GetOrgProvisionResponse(ctx)
	if err != nil {
		return nil, err
	}

	serverInfo, err := s.WaypointRepo.GetVersionInfo(ctx)
	if err != nil {
		return nil, err
	}

	runnerInfo, err := s.WaypointRepo.GetRunner(ctx, orgID)
	if err != nil {
		return nil, err
	}

	fmt.Println(workflowResp, runnerInfo, serverInfo)
	return nil, nil
}
